// Package config unifies the environment configuration owned by every platform module.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"

	"github.com/pixelados-net/asset-cli/platform/arcturus"
	"github.com/pixelados-net/asset-cli/platform/logger"
	"github.com/pixelados-net/asset-cli/platform/minio"
	"github.com/pixelados-net/asset-cli/platform/pixels"
	"github.com/pixelados-net/asset-cli/platform/redis"
)

// EnvironmentPrefix namespaces every asset-cli environment variable.
const EnvironmentPrefix = "ASSET_CLI_"

// EmulatorKind selects which emulator backend the sync realm targets.
type EmulatorKind string

const (
	// EmulatorArcturus targets the Arcturus MySQL schema.
	EmulatorArcturus EmulatorKind = "arcturus"
	// EmulatorPixels targets the Pixels PostgreSQL schema.
	EmulatorPixels EmulatorKind = "pixels"
)

// UnmarshalText validates and stores an emulator kind.
func (kind *EmulatorKind) UnmarshalText(value []byte) error {
	parsed := EmulatorKind(value)
	if parsed != EmulatorArcturus && parsed != EmulatorPixels {
		return fmt.Errorf("unsupported emulator kind %q", value)
	}
	*kind = parsed
	return nil
}

// Config aggregates every platform module's own configuration struct.
type Config struct {
	// Logger holds the process logger settings.
	Logger logger.Config `envPrefix:"LOG_"`
	// MinIO holds the object storage settings.
	MinIO minio.Config `envPrefix:"MINIO_"`
	// Redis holds the cache/cursor settings used by the sync realm.
	Redis redis.Config `envPrefix:"REDIS_"`
	// Emulator selects which backend the sync realm targets.
	Emulator EmulatorKind `env:"EMULATOR_KIND" envDefault:"arcturus"`
	// Arcturus holds the Arcturus MySQL connection settings.
	Arcturus arcturus.Config `envPrefix:"ARCTURUS_"`
	// Pixels holds the Pixels PostgreSQL connection settings.
	Pixels pixels.Config `envPrefix:"PIXELS_"`
}

// Load parses every platform module's configuration from ASSET_CLI_* variables.
func Load() (Config, error) {
	config, err := env.ParseAsWithOptions[Config](env.Options{Prefix: EnvironmentPrefix})
	if err != nil {
		return Config{}, err
	}
	if err := config.MinIO.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}
