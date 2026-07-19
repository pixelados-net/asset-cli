package pixels

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client wraps the reusable Pixels PostgreSQL driver.
type Client struct {
	pool *pgxpool.Pool
}

// New creates a Pixels PostgreSQL client without performing network I/O.
func New(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, config.SSLMode)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return &Client{pool: pool}, nil
}

// Pool returns the underlying PostgreSQL connection pool.
func (client *Client) Pool() *pgxpool.Pool {
	return client.pool
}

// Ping verifies the Pixels connection.
func (client *Client) Ping(ctx context.Context) error {
	return client.pool.Ping(ctx)
}

// Close closes the Pixels connection pool.
func (client *Client) Close() {
	client.pool.Close()
}
