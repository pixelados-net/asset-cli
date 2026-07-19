package stats

import (
	"context"
	"sort"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// maxConcurrentCounts caps concurrent MinIO listing requests issued by Nitro; each
// category can hold tens of thousands of objects, so counting categories concurrently
// shortens wall-clock time instead of enumerating every category one at a time.
const maxConcurrentCounts = 4

type service struct {
	storage Storage
}

// NewService creates the stats realm's service backed by the injected MinIO client.
func NewService(storage *minio.Client) Service {
	return newService(storage)
}

func newService(storage Storage) *service {
	return &service{storage: storage}
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
