// Package effects cross-checks avatar effect bundles against EffectMap.json.
package effects

import "context"

// Service defines the effects realm's capabilities, independent of transport.
type Service interface {
	// Check cross-references effect bundles against client-declared libraries.
	Check(ctx context.Context) (Report, error)
}

// BundleStorage is the object-storage subset used by Check.
type BundleStorage interface {
	// ListNames returns every effect bundle name stored in the bucket.
	ListNames(ctx context.Context) ([]string, error)
}

// Catalog is the client-gamedata subset used by Check.
type Catalog interface {
	// ListNames returns every effect library declared by EffectMap.json.
	ListNames(ctx context.Context) ([]string, error)
}

// Report is the result of an effect bundle cross-check.
type Report struct {
	// Matched is the number of names present in the bucket and EffectMap.json.
	Matched int
	// Orphaned lists bundles with no EffectMap.json library.
	Orphaned []string
	// Missing lists EffectMap.json libraries with no bundle.
	Missing []string
}

// OK reports whether every declared effect library has a bundle.
func (report Report) OK() bool { return len(report.Missing) == 0 }
