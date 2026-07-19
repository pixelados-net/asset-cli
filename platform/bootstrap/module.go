// Package bootstrap composes package-owned modules into the process Fx graph.
package bootstrap

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/config"
	"github.com/pixelados-net/asset-cli/platform/logger"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

// Module composes the platform-owned modules shared by every realm. It intentionally
// declares no domain (internal/) providers: realms supply their own Fx module to
// Invoke, which keeps this package free of import cycles back into internal/.
var Module = fx.Module("bootstrap",
	config.Module,
	logger.Module,
	minio.Module,
)
