package clothing

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/internal/clothing/gamedata"
)

// Module provides the clothing realm and its FigureMap.json catalog.
var Module = fx.Module("clothing", fx.Provide(
	NewService,
	fx.Annotate(gamedata.New, fx.As(new(Catalog))),
))
