package structure

import (
	"context"
	"sort"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

// maxConcurrentChecks caps concurrent MinIO requests issued by Check and Create;
// each check is I/O-bound, so bounded concurrency shortens wall-clock time on
// buckets holding tens of thousands of objects without opening unbounded connections.
const maxConcurrentChecks = 8

type service struct {
	storage Storage
}

// NewService creates the structure realm's service backed by the injected MinIO client.
func NewService(storage *minio.Client) Service {
	return newService(storage)
}

func newService(storage Storage) *service {
	return &service{storage: storage}
}

func (svc *service) Check(ctx context.Context) (Report, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentChecks)
	var mutex sync.Mutex
	var report Report

	for _, path := range ExpectedPaths {
		group.Go(func() error {
			exists, err := svc.storage.Exists(groupCtx, path)
			if err != nil {
				return err
			}
			mutex.Lock()
			if exists {
				report.Present = append(report.Present, path)
			} else {
				report.Missing = append(report.Missing, path)
			}
			mutex.Unlock()
			return nil
		})
	}
	for _, path := range FlatPaths {
		group.Go(func() error {
			nested, err := svc.storage.SubPrefixes(groupCtx, path)
			if err != nil {
				return err
			}
			if len(nested) == 0 {
				return nil
			}
			mutex.Lock()
			report.Nested = append(report.Nested, nested...)
			mutex.Unlock()
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return Report{}, err
	}
	sort.Strings(report.Present)
	sort.Strings(report.Missing)
	sort.Strings(report.Nested)
	return report, nil
}

func (svc *service) Create(ctx context.Context) ([]string, error) {
	report, err := svc.Check(ctx)
	if err != nil {
		return nil, err
	}

	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(maxConcurrentChecks)
	var mutex sync.Mutex
	var created []string

	for _, path := range report.Missing {
		if !strings.HasSuffix(path, "/") {
			continue // an exact file key is real content to upload, not a placeholder folder
		}
		group.Go(func() error {
			if err := svc.storage.Touch(groupCtx, path+placeholderSuffix); err != nil {
				return err
			}
			mutex.Lock()
			created = append(created, path)
			mutex.Unlock()
			return nil
		})
	}

	waitErr := group.Wait()
	sort.Strings(created)
	if waitErr != nil {
		return created, waitErr
	}
	return created, nil
}
