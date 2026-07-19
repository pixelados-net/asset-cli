package furniture

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

// NewService creates the furniture realm's service backed by the injected MinIO
// client and furniture catalog.
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
		names, err := svc.bundles.ListClassnames(groupCtx)
		if err != nil {
			return err
		}
		bundleNames = names
		return nil
	})
	group.Go(func() error {
		names, err := svc.catalog.ListClassnames(groupCtx)
		if err != nil {
			return err
		}
		catalogNames = names
		return nil
	})
	if err := group.Wait(); err != nil {
		return Report{}, err
	}

	bundleSet := toSet(bundleNames)
	catalogSet := toSet(catalogNames)

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

func toSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}
	return set
}
