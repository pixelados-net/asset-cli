package stats

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/pixelados-net/asset-cli/internal/clothing"
	"github.com/pixelados-net/asset-cli/internal/effects"
	"github.com/pixelados-net/asset-cli/internal/furniture"
	"github.com/pixelados-net/asset-cli/internal/pets"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

// maxConcurrentCounts caps concurrent MinIO listing requests issued by Nitro; each
// category can hold tens of thousands of objects, so counting categories concurrently
// shortens wall-clock time instead of enumerating every category one at a time.
const maxConcurrentCounts = 4

// maxConcurrentOrphanChecks caps category checks run by Orphans. Each check
// performs independent object-storage reads and can safely overlap.
const maxConcurrentOrphanChecks = 4

// ClothingChecker is the clothing check subset stats consumes.
type ClothingChecker interface {
	Check(ctx context.Context) (clothing.Report, error)
}

// EffectsChecker is the effects check subset stats consumes.
type EffectsChecker interface {
	Check(ctx context.Context) (effects.Report, error)
}

// FurnitureChecker is the subset of the furniture realm's port stats needs. It
// matches furniture.Service's shape structurally, so the real furniture.Service
// satisfies it without stats importing anything beyond furniture.Report.
type FurnitureChecker interface {
	Check(ctx context.Context) (furniture.Report, error)
}

// PetsChecker is the pets check subset stats consumes.
type PetsChecker interface {
	Check(ctx context.Context) (pets.Report, error)
}

type service struct {
	storage   Storage
	clothing  ClothingChecker
	effects   EffectsChecker
	furniture FurnitureChecker
	pets      PetsChecker
}

// NewService creates stats backed by MinIO and every category-owned cross-check.
func NewService(storage *minio.Client, clothingChecker clothing.Service, effectsChecker effects.Service,
	furnitureChecker furniture.Service, petsChecker pets.Service,
) Service {
	return newService(storage, clothingChecker, effectsChecker, furnitureChecker, petsChecker)
}

func newService(storage Storage, clothingChecker ClothingChecker, effectsChecker EffectsChecker,
	furnitureChecker FurnitureChecker, petsChecker PetsChecker,
) *service {
	return &service{storage: storage, clothing: clothingChecker, effects: effectsChecker,
		furniture: furnitureChecker, pets: petsChecker}
}

func (svc *service) Nitro(ctx context.Context) ([]Count, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentCounts)
	var mutex sync.Mutex
	counts := make([]Count, 0, len(NitroCategories))

	for _, category := range NitroCategories {
		group.Go(func() error {
			total, err := svc.storage.CountByExtension(groupCtx, category.Path, nitroExtension)
			if err != nil {
				return err
			}
			mutex.Lock()
			counts = append(counts, Count{Name: category.Name, Total: total})
			mutex.Unlock()
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].Name < counts[j].Name })
	return counts, nil
}

func (svc *service) Orphans(ctx context.Context) ([]OrphanReport, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentOrphanChecks)
	var mutex sync.Mutex
	reports := make([]OrphanReport, 0, maxConcurrentOrphanChecks)
	appendReport := func(report OrphanReport) {
		mutex.Lock()
		reports = append(reports, report)
		mutex.Unlock()
	}
	group.Go(func() error {
		report, err := svc.clothing.Check(groupCtx)
		if err == nil {
			appendReport(OrphanReport{Category: "clothing", Matched: report.Matched,
				Orphaned: len(report.Orphaned), Missing: len(report.Missing)})
		}
		return err
	})
	group.Go(func() error {
		report, err := svc.effects.Check(groupCtx)
		if err == nil {
			appendReport(OrphanReport{Category: "effects", Matched: report.Matched,
				Orphaned: len(report.Orphaned), Missing: len(report.Missing)})
		}
		return err
	})
	group.Go(func() error {
		report, err := svc.furniture.Check(groupCtx)
		if err == nil {
			appendReport(OrphanReport{Category: "furniture", Matched: report.Matched,
				Orphaned: len(report.Orphaned), Missing: len(report.Missing)})
		}
		return err
	})
	group.Go(func() error {
		report, err := svc.pets.Check(groupCtx)
		if err == nil {
			appendReport(OrphanReport{Category: "pets", Matched: report.Matched,
				Orphaned: len(report.Orphaned), Missing: len(report.Missing)})
		}
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	sort.Slice(reports, func(i, j int) bool { return reports[i].Category < reports[j].Category })
	return reports, nil
}
