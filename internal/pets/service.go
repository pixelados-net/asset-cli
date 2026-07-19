package pets

import (
	"context"
	"sort"

	"golang.org/x/sync/errgroup"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

type service struct {
	bundles BundleStorage
	catalog Catalog
}

// NewService creates the pets realm's service.
func NewService(client *minio.Client, catalog Catalog) Service {
	return newService(&bundleStorage{client: client}, catalog)
}

func newService(bundles BundleStorage, catalog Catalog) *service {
	return &service{bundles: bundles, catalog: catalog}
}

func (svc *service) Check(ctx context.Context) (Report, error) {
	group, groupCtx := errgroup.WithContext(ctx)
	var bundleNames, catalogNames []string
	group.Go(func() error {
		var err error
		bundleNames, err = svc.bundles.ListNames(groupCtx)
		return err
	})
	group.Go(func() error {
		var err error
		catalogNames, err = svc.catalog.ListNames(groupCtx)
		return err
	})
	if err := group.Wait(); err != nil {
		return Report{}, err
	}

	bundleSet := nameSet(bundleNames)
	catalogSet := nameSet(catalogNames)
	var report Report
	for name := range bundleSet {
		if catalogSet[name] {
			report.Matched++
		} else {
			report.Orphaned = append(report.Orphaned, name)
		}
	}
	for name := range catalogSet {
		if !bundleSet[name] {
			report.Missing = append(report.Missing, name)
		}
	}
	sort.Strings(report.Orphaned)
	sort.Strings(report.Missing)
	return report, nil
}

func nameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		if name != "" {
			set[name] = true
		}
	}
	return set
}
