// Package structure verifies and repairs the asset bucket's expected folder layout.
package structure

import "context"

// Service defines the structure realm's capabilities, independent of transport.
type Service interface {
	// Check reports which expected bucket paths exist and which are missing.
	Check(ctx context.Context) (Report, error)
	// Create adds a placeholder object for every expected path that is missing
	// and returns the paths it created.
	Create(ctx context.Context) ([]string, error)
}

// Storage is the subset of object storage operations the structure realm needs.
type Storage interface {
	// Exists reports whether any object exists under the given key prefix.
	Exists(ctx context.Context, prefix string) (bool, error)
	// Touch creates an empty placeholder object at key.
	Touch(ctx context.Context, key string) error
}

// Report is the result of a structure integrity check.
type Report struct {
	// Present lists expected paths already found in the bucket.
	Present []string
	// Missing lists expected paths not found in the bucket.
	Missing []string
}

// OK reports whether every expected path is present.
func (report Report) OK() bool {
	return len(report.Missing) == 0
}
