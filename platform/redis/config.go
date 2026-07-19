// Package redis contains the reusable Redis cache/cursor adapter.
package redis

import "time"

// Config contains Redis connection settings.
type Config struct {
	// Address is the Redis server address.
	Address string `env:"ADDR" envDefault:"127.0.0.1:6379"`
	// Username is the Redis ACL username.
	Username string `env:"USERNAME" envDefault:""`
	// Password is the Redis password.
	Password string `env:"PASSWORD" envDefault:""`
	// Database is the selected Redis database.
	Database int `env:"DB" envDefault:"0"`
	// DialTimeout limits connection establishment.
	DialTimeout time.Duration `env:"DIAL_TIMEOUT" envDefault:"5s"`
	// HealthTimeout limits health probes.
	HealthTimeout time.Duration `env:"HEALTH_TIMEOUT" envDefault:"2s"`
}
