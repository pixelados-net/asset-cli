package arcturus

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Client wraps the reusable Arcturus MySQL driver.
type Client struct {
	db *sql.DB
}

// New creates an Arcturus MySQL client without performing network I/O.
func New(config Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", config.User, config.Password, config.Host, config.Port, config.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

// DB returns the underlying MySQL connection pool.
func (client *Client) DB() *sql.DB {
	return client.db
}

// Ping verifies the Arcturus connection.
func (client *Client) Ping(ctx context.Context) error {
	return client.db.PingContext(ctx)
}

// Close closes the Arcturus connection pool.
func (client *Client) Close() error {
	return client.db.Close()
}
