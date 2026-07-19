package sync

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/minio"
	"github.com/pixelados-net/asset-cli/platform/redis"
)

// Module provides the sync realm's service and its client catalog (Redis-cached).
// It does not provide an EmulatorCatalog: the CLI picks exactly one of
// ArcturusModule or PixelsModule based on ASSET_CLI_EMULATOR_KIND and composes it
// alongside this Module, so only the configured backend is ever constructed.
var Module = fx.Module("sync",
	redis.Module,
	fx.Provide(
		NewService,
		provideClientCatalog,
	),
)

func provideClientCatalog(client *minio.Client, redisClient *redis.Client) ClientCatalog {
	return NewCachedClientCatalog(NewClientCatalog(client), redisClient)
}
