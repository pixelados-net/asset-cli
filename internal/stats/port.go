// Package stats reports content counts for the asset-cli bucket.
package stats

import "context"

// Service defines the stats realm's capabilities, independent of transport.
type Service interface {
	// Nitro counts .nitro bundles per content category.
	Nitro(ctx context.Context) ([]Count, error)
	// Orphans summarizes bundle-vs-catalog cross-checks per content category.
	// Furniture is the only category wired in today; clothing, effects, and pets
	// follow the same shape once their own realms grow a Catalog cross-check.
	Orphans(ctx context.Context) ([]OrphanReport, error)
}

// Storage is the subset of object storage operations the stats realm needs.
type Storage interface {
	// CountByExtension counts objects under prefix whose key ends with extension.
	CountByExtension(ctx context.Context, prefix, extension string) (int, error)
}

// Count is the number of bundles found for one content category.
type Count struct {
	// Name is the category label.
	Name string
	// Total is the number of .nitro bundles found.
	Total int
}

// OrphanReport summarizes one content category's bundle-vs-catalog cross-check.
type OrphanReport struct {
	// Category names the content category (e.g. "furniture").
	Category string
	// Matched is the number of classnames present in both the bucket and the catalog.
	Matched int
	// Orphaned is the number of classnames with a bundle file but no catalog entry.
	Orphaned int
	// Missing is the number of classnames with a catalog entry but no bundle file.
	Missing int
}
