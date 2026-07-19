// Package pets cross-checks pet bundles against Nitro's pet type assets.
package pets

import "context"

// Service defines the pets realm's capabilities, independent of transport.
type Service interface {
	// Check cross-references pet bundles against client-required asset names.
	Check(ctx context.Context) (Report, error)
}

// BundleStorage is the object-storage subset used by Check.
type BundleStorage interface {
	// ListNames returns every pet bundle name stored in the bucket.
	ListNames(ctx context.Context) ([]string, error)
}

// Catalog provides the pet asset names required by the Nitro client.
type Catalog interface {
	// ListNames returns the configured pet type asset names.
	ListNames(ctx context.Context) ([]string, error)
}

// Report is the result of a pet bundle cross-check.
type Report struct {
	// Matched is the number of names present in the bucket and pet type list.
	Matched int
	// Orphaned lists bundles not referenced by a standard pet type.
	Orphaned []string
	// Missing lists required pet type assets with no bundle.
	Missing []string
}

// OK reports whether every required pet type asset has a bundle.
func (report Report) OK() bool { return len(report.Missing) == 0 }
