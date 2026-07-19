package stats

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/internal/clothing"
	"github.com/pixelados-net/asset-cli/internal/effects"
	"github.com/pixelados-net/asset-cli/internal/furniture"
	"github.com/pixelados-net/asset-cli/internal/pets"
)

// Module provides stats and composes every category-owned check it summarizes.
var Module = fx.Module("stats",
	clothing.Module,
	effects.Module,
	furniture.Module,
	pets.Module,
	fx.Provide(NewService),
)
