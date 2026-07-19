package stats

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/pixelados-net/asset-cli/internal/furniture"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

// maxConcurrentCounts caps concurrent MinIO listing requests issued by Nitro; each
// category can hold tens of thousands of objects, so counting categories concurrently
// shortens wall-clock time instead of enumerating every category one at a time.
const maxConcurrentCounts = 4

// FurnitureChecker is the subset of the furniture realm's port stats needs. It
// matches furniture.Service's shape structurally, so the real furniture.Service
// satisfies it without stats importing anything beyond furniture.Report.
type FurnitureChecker interface {
	Check(ctx context.Context) (furniture.Report, error)
}

type service struct {
	storage   Storage
	furniture FurnitureChecker
}

// NewService creates the stats realm's service backed by the injected MinIO client
// and the furniture realm's cross-check.
func NewService(storage *minio.Client, furnitureChecker furniture.Service) Service {
	return newService(storage, furnitureChecker)
}

func newService(storage Storage, furnitureChecker FurnitureChecker) *service {
	return &service{storage: storage, furniture: furnitureChecker}
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
	report, err := svc.furniture.Check(ctx)
	if err != nil {
		return nil, err
	}
	return []OrphanReport{
		{
			Category: "furniture",
			Matched:  report.Matched,
			Orphaned: len(report.Orphaned),
			Missing:  len(report.Missing),
		},
	}, nil
}
