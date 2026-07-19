// Package clothing cross-checks avatar clothing bundles against FigureMap.json.
package clothing

import "context"

// Service defines the clothing realm's capabilities, independent of transport.
type Service interface {
	// Check cross-references clothing bundles against client-declared libraries.
	Check(ctx context.Context) (Report, error)
}

// BundleStorage is the object-storage subset used by Check.
type BundleStorage interface {
	// ListNames returns every clothing bundle name stored in the bucket.
	ListNames(ctx context.Context) ([]string, error)
}

// Catalog is the client-gamedata subset used by Check.
type Catalog interface {
	// ListNames returns every clothing library declared by FigureMap.json.
	ListNames(ctx context.Context) ([]string, error)
}

// Report is the result of a clothing bundle cross-check.
type Report struct {
	// Matched is the number of names present in the bucket and FigureMap.json.
	Matched int
	// Orphaned lists bundles with no FigureMap.json library.
	Orphaned []string
	// Missing lists FigureMap.json libraries with no bundle.
	Missing []string
}

// OK reports whether every declared clothing library has a bundle.
func (report Report) OK() bool { return len(report.Missing) == 0 }
