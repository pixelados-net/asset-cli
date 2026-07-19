package sync

import (
	"context"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/pixelados-net/asset-cli/platform/redis"
)

// maxConcurrentWrites caps concurrent emulator batch operations; each batch is
// I/O-bound, so bounded concurrency shortens wall-clock time on catalogs holding
// tens of thousands of definitions without opening unbounded connections.
const maxConcurrentWrites = 8

// writeBatchSize is how many definitions Apply inserts or updates per emulator
// round trip.
const writeBatchSize = 500

// cursorStore persists Apply's insert resume point. It is an optimization only:
// Apply always recomputes the missing set from the emulator's current state, so
// a stale, missing, or wrong cursor value never causes an incorrect result, only
// redundant work on the next run. Updates need no cursor: overwriting the same
// row with the same target value twice is harmless, unlike Arcturus's
// unconstrained item_name column, which can accumulate duplicate inserts.
type cursorStore interface {
	Get(ctx context.Context) (string, error)
	Set(ctx context.Context, classname string) error
	Clear(ctx context.Context) error
}

type service struct {
	client   ClientCatalog
	emulator EmulatorCatalog
	cursor   cursorStore
}

// NewService creates the sync realm's service backed by the injected Redis client
// for Apply's insert resume cursor and the client catalog's cache.
func NewService(client ClientCatalog, emulator EmulatorCatalog, redisClient *redis.Client) Service {
	return newService(client, emulator, newCursor(redisClient))
}

func newService(client ClientCatalog, emulator EmulatorCatalog, cursor cursorStore) *service {
	return &service{client: client, emulator: emulator, cursor: cursor}
}

// baseClassname strips a trailing "*N" color-index suffix, same rule as the
// furniture realm's diff — a color variant shares its base classname's bundle
// and, here, its base classname's definition row.
func baseClassname(classname string) string {
	if index := strings.IndexByte(classname, '*'); index >= 0 {
		return classname[:index]
	}
	return classname
}

func (svc *service) Check(ctx context.Context) (Report, error) {
	clientByName, emulatorByName, err := svc.listBoth(ctx)
	if err != nil {
		return Report{}, err
	}

	var report Report
	for name := range clientByName {
		if _, ok := emulatorByName[name]; !ok {
			report.Missing = append(report.Missing, name)
		}
	}
	for name := range emulatorByName {
		if _, ok := clientByName[name]; !ok {
			report.Orphaned = append(report.Orphaned, name)
		}
	}
	for _, change := range nameChanges(clientByName, emulatorByName) {
		report.NameChanges = append(report.NameChanges, change)
	}

	sort.Strings(report.Missing)
	sort.Strings(report.Orphaned)
	sort.Slice(report.NameChanges, func(i, j int) bool { return report.NameChanges[i].Classname < report.NameChanges[j].Classname })
	return report, nil
}

func (svc *service) listBoth(ctx context.Context) (map[string]Definition, map[string]Definition, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	var clientDefs, emulatorDefs []Definition

	group.Go(func() error {
		defs, err := svc.client.ListDefinitions(groupCtx)
		if err != nil {
			return err
		}
		clientDefs = defs
		return nil
	})
	group.Go(func() error {
		defs, err := svc.emulator.ListDefinitions(groupCtx)
		if err != nil {
			return err
		}
		emulatorDefs = defs
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, nil, err
	}
	return indexByBaseClassname(clientDefs), indexByBaseClassname(emulatorDefs), nil
}

// nameChanges finds classnames present on both sides whose public_name or
// description differs, keyed by classname for easy lookup during Apply.
func nameChanges(clientByName, emulatorByName map[string]Definition) map[string]NameChange {
	changes := make(map[string]NameChange)
	for name, clientDef := range clientByName {
		emulatorDef, ok := emulatorByName[name]
		if !ok {
			continue
		}
		if clientDef.PublicName != emulatorDef.PublicName || clientDef.Description != emulatorDef.Description {
			changes[name] = NameChange{
				Classname:           name,
				ClientName:          clientDef.PublicName,
				EmulatorName:        emulatorDef.PublicName,
				ClientDescription:   clientDef.Description,
				EmulatorDescription: emulatorDef.Description,
			}
		}
	}
	return changes
}

func (svc *service) Apply(ctx context.Context) (ApplyResult, error) {
	clientByName, emulatorByName, err := svc.listBoth(ctx)
	if err != nil {
		return ApplyResult{}, err
	}

	missing := make([]Definition, 0, len(clientByName))
	for name, definition := range clientByName {
		if _, ok := emulatorByName[name]; !ok {
			missing = append(missing, definition)
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].Classname < missing[j].Classname })

	changedNames := nameChanges(clientByName, emulatorByName)
	changed := make([]Definition, 0, len(changedNames))
	for name := range changedNames {
		changed = append(changed, clientByName[name])
	}
	sort.Slice(changed, func(i, j int) bool { return changed[i].Classname < changed[j].Classname })

	created, err := svc.insertMissing(ctx, missing)
	if err != nil {
		return ApplyResult{Created: created}, err
	}
	updated, err := svc.updateChanged(ctx, changed)
	if err != nil {
		return ApplyResult{Created: created, Updated: updated}, err
	}
	return ApplyResult{Created: created, Updated: updated}, nil
}

func (svc *service) insertMissing(ctx context.Context, missing []Definition) ([]string, error) {
	if cursorValue, cursorErr := svc.cursor.Get(ctx); cursorErr == nil && cursorValue != "" {
		index := sort.Search(len(missing), func(i int) bool { return missing[i].Classname > cursorValue })
		missing = missing[index:]
	}
	if len(missing) == 0 {
		_ = svc.cursor.Clear(ctx)
		return nil, nil
	}

	batchCount := (len(missing) + writeBatchSize - 1) / writeBatchSize
	completed := make([]bool, batchCount)
	var mutex sync.Mutex

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentWrites)
	for batchIndex := 0; batchIndex < batchCount; batchIndex++ {
		start := batchIndex * writeBatchSize
		end := start + writeBatchSize
		if end > len(missing) {
			end = len(missing)
		}
		batch := missing[start:end]
		index := batchIndex
		group.Go(func() error {
			if err := svc.emulator.InsertDefinitions(groupCtx, batch); err != nil {
				return err
			}
			mutex.Lock()
			completed[index] = true
			mutex.Unlock()
			return nil
		})
	}
	writeErr := group.Wait()

	contiguous := 0
	for contiguous < batchCount && completed[contiguous] {
		contiguous++
	}
	if contiguous == batchCount {
		_ = svc.cursor.Clear(ctx)
	} else if contiguous > 0 {
		_ = svc.cursor.Set(ctx, missing[contiguous*writeBatchSize-1].Classname)
	}
	if writeErr != nil {
		return nil, writeErr
	}

	created := make([]string, len(missing))
	for i, definition := range missing {
		created[i] = definition.Classname
	}
	sort.Strings(created)
	return created, nil
}

func (svc *service) updateChanged(ctx context.Context, changed []Definition) ([]string, error) {
	if len(changed) == 0 {
		return nil, nil
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentWrites)
	for start := 0; start < len(changed); start += writeBatchSize {
		end := start + writeBatchSize
		if end > len(changed) {
			end = len(changed)
		}
		batch := changed[start:end]
		group.Go(func() error {
			return svc.emulator.UpdateDefinitions(groupCtx, batch)
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}

	updated := make([]string, len(changed))
	for i, definition := range changed {
		updated[i] = definition.Classname
	}
	sort.Strings(updated)
	return updated, nil
}

func indexByBaseClassname(definitions []Definition) map[string]Definition {
	index := make(map[string]Definition, len(definitions))
	for _, definition := range definitions {
		name := baseClassname(definition.Classname)
		if _, exists := index[name]; !exists {
			index[name] = definition
		}
	}
	return index
}
