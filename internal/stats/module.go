package stats

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/internal/furniture"
)

// Module provides the stats realm's service from the injected MinIO client and
// composes furniture.Module so Orphans can resolve furniture.Service. As more
// categories grow their own catalog cross-check, compose their modules here too.
var Module = fx.Module("stats", furniture.Module, fx.Provide(NewService))
