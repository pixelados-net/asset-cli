package config

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/arcturus"
	"github.com/pixelados-net/asset-cli/platform/logger"
	"github.com/pixelados-net/asset-cli/platform/minio"
	"github.com/pixelados-net/asset-cli/platform/pixels"
	"github.com/pixelados-net/asset-cli/platform/redis"
)

// Module provides the aggregate Config and every platform module's own struct.
var Module = fx.Module("config", fx.Provide(
	Load,
	provideLoggerConfig,
	provideMinIOConfig,
	provideRedisConfig,
	provideArcturusConfig,
	providePixelsConfig,
))

func provideLoggerConfig(config Config) logger.Config {
	return config.Logger
}

func provideMinIOConfig(config Config) minio.Config {
	return config.MinIO
}

func provideRedisConfig(config Config) redis.Config {
	return config.Redis
}

func provideArcturusConfig(config Config) arcturus.Config {
	return config.Arcturus
}

func providePixelsConfig(config Config) pixels.Config {
	return config.Pixels
}
