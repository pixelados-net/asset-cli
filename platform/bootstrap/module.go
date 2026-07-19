// Package bootstrap composes package-owned modules into the process Fx graph.
package bootstrap

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/config"
	"github.com/pixelados-net/asset-cli/platform/logger"
	"github.com/pixelados-net/asset-cli/platform/minio"
)

// Module composes package-owned modules without declaring domain providers.
var Module = fx.Module("bootstrap",
	config.Module,
	logger.Module,
	minio.Module,
)
