// Package stats reports content counts for the asset-cli bucket.
package stats

import "context"

// Service defines the stats realm's capabilities, independent of transport.
type Service interface {
	// Nitro counts .nitro bundles per content category.
	Nitro(ctx context.Context) ([]Count, error)
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
