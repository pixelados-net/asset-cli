package pixels

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// rows is the subset of pgx.Rows Adapter needs.
type rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

// pgxPool adapts *pgxpool.Pool to the pool interface Adapter depends on.
type pgxPool struct {
	pool *pgxpool.Pool
}

func (wrapper pgxPool) Query(ctx context.Context, sql string, args ...any) (rows, error) {
	return wrapper.pool.Query(ctx, sql, args...)
}

func (wrapper pgxPool) Exec(ctx context.Context, sql string, args ...any) error {
	_, err := wrapper.pool.Exec(ctx, sql, args...)
	return err
}
