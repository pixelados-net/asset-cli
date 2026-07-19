package arcturus

import (
	"context"
	"database/sql"
)

// rows is the subset of *sql.Rows Adapter needs.
type rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

// querier is the subset of *sql.DB Adapter needs.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (rows, error)
	ExecContext(ctx context.Context, query string, args ...any) error
}

// sqlDB adapts *sql.DB to the querier interface Adapter depends on.
type sqlDB struct {
	db *sql.DB
}

func (wrapper sqlDB) QueryContext(ctx context.Context, query string, args ...any) (rows, error) {
	return wrapper.db.QueryContext(ctx, query, args...)
}

func (wrapper sqlDB) ExecContext(ctx context.Context, query string, args ...any) error {
	_, err := wrapper.db.ExecContext(ctx, query, args...)
	return err
}
