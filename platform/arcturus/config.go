// Package arcturus contains the reusable Arcturus MySQL adapter.
package arcturus

import "fmt"

// Config contains Arcturus MySQL connection settings.
type Config struct {
	// Host is the MySQL server host.
	Host string `env:"HOST" envDefault:"127.0.0.1"`
	// Port is the MySQL server port.
	Port int `env:"PORT" envDefault:"3306"`
	// Database is the Arcturus database name.
	Database string `env:"DATABASE" envDefault:""`
	// User authenticates against the MySQL server.
	User string `env:"USER" envDefault:"root"`
	// Password authenticates against the MySQL server.
	Password string `env:"PASSWORD" envDefault:""`
}

// Validate reports whether the mandatory Arcturus settings were supplied.
func (config Config) Validate() error {
	if config.Database == "" {
		return fmt.Errorf("arcturus database is required")
	}
	return nil
}
