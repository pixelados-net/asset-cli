// Package pixels contains the reusable Pixels PostgreSQL adapter.
package pixels

import "fmt"

// Config contains Pixels PostgreSQL connection settings.
type Config struct {
	// Host is the PostgreSQL server host.
	Host string `env:"HOST" envDefault:"127.0.0.1"`
	// Port is the PostgreSQL server port.
	Port int `env:"PORT" envDefault:"5432"`
	// Database is the Pixels database name.
	Database string `env:"DATABASE" envDefault:""`
	// User authenticates against the PostgreSQL server.
	User string `env:"USER" envDefault:""`
	// Password authenticates against the PostgreSQL server.
	Password string `env:"PASSWORD" envDefault:""`
	// SSLMode selects the PostgreSQL SSL negotiation mode.
	SSLMode string `env:"SSL_MODE" envDefault:"disable"`
}

// Validate reports whether the mandatory Pixels settings were supplied.
func (config Config) Validate() error {
	if config.Database == "" {
		return fmt.Errorf("pixels database is required")
	}
	if config.User == "" {
		return fmt.Errorf("pixels user is required")
	}
	return nil
}
