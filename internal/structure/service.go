package structure

import (
	"context"
	"sort"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

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
	var report Report
	for _, path := range ExpectedPaths {
		exists, err := svc.storage.Exists(ctx, path)
		if err != nil {
			return Report{}, err
		}
		if exists {
			report.Present = append(report.Present, path)
		} else {
			report.Missing = append(report.Missing, path)
		}
	}
	sort.Strings(report.Present)
	sort.Strings(report.Missing)
	return report, nil
}

func (svc *service) Create(ctx context.Context) ([]string, error) {
	report, err := svc.Check(ctx)
	if err != nil {
		return nil, err
	}
	created := make([]string, 0, len(report.Missing))
	for _, path := range report.Missing {
		if err := svc.storage.Touch(ctx, path+placeholderSuffix); err != nil {
			return created, err
		}
		created = append(created, path)
	}
	return created, nil
}
