// Package config unifies the environment configuration owned by every platform module.
package config

import (
	"github.com/caarlos0/env/v11"

	"github.com/pixelados-net/asset-cli/platform/logger"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

// EnvironmentPrefix namespaces every asset-cli environment variable.
const EnvironmentPrefix = "ASSET_CLI_"

// Config aggregates every platform module's own configuration struct.
type Config struct {
	// Logger holds the process logger settings.
	Logger logger.Config `envPrefix:"LOG_"`
	// MinIO holds the object storage settings.
	MinIO minio.Config `envPrefix:"MINIO_"`
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
