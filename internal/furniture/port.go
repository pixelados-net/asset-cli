// Package furniture cross-checks furniture bundles against the furniture catalog.
package furniture

import "context"

// Service defines the furniture realm's capabilities, independent of transport.
type Service interface {
	// Check cross-references bundle files under furniture/bundles/ against the
	// furniture catalog's declared classnames.
	Check(ctx context.Context) (Report, error)
}

// BundleStorage is the subset of object storage operations the furniture realm needs.
type BundleStorage interface {
	// ListClassnames returns every classname with a bundle under furniture/bundles/.
	ListClassnames(ctx context.Context) ([]string, error)
}

// Catalog is what Check needs from wherever the furniture catalog lives. A gamedata
// JSON adapter satisfies this today; a future SQL or Mongo adapter can satisfy the
// same interface without changing Service.
type Catalog interface {
	// ListClassnames returns every classname the catalog declares.
	ListClassnames(ctx context.Context) ([]string, error)
}

// Report is the result of a furniture catalog cross-check.
type Report struct {
	// Matched is the number of classnames present in both the bucket and the catalog.
	Matched int
	// Orphaned lists classnames with a bundle file but no catalog entry.
	Orphaned []string
	// Missing lists classnames with a catalog entry but no bundle file.
	Missing []string
}

// OK reports whether every catalog entry has a matching bundle file. An orphaned
// bundle alone does not fail the check: it may be an intentionally kept file for
// existing rooms, whereas a missing bundle is a live 404 for players.
func (report Report) OK() bool {
	return len(report.Missing) == 0
}
