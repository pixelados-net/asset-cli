package furniture

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/internal/furniture/gamedata"
)

// Module provides the furniture realm's service, backed by the gamedata JSON
// catalog. Swapping to a future SQL/Mongo-backed catalog only means replacing this
// provider with one annotated fx.As(new(Catalog)); Service never changes.
var Module = fx.Module("furniture", fx.Provide(
	NewService,
	fx.Annotate(gamedata.New, fx.As(new(Catalog))),
))
