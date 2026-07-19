package config

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/logger"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

// Module provides the aggregate Config and every platform module's own struct.
var Module = fx.Module("config", fx.Provide(Load, provideLoggerConfig, provideMinIOConfig))

func provideLoggerConfig(config Config) logger.Config {
	return config.Logger
}

func provideMinIOConfig(config Config) minio.Config {
	return config.MinIO
}
