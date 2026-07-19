package effects

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/internal/effects/gamedata"
)

// Module provides the effects realm and its EffectMap.json catalog.
var Module = fx.Module("effects", fx.Provide(
	NewService,
	fx.Annotate(gamedata.New, fx.As(new(Catalog))),
))
