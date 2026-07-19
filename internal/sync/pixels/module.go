package pixels

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/pixels"
)

// Module composes the Pixels platform client and provides it as the sync
// realm's EmulatorCatalog.
var Module = fx.Module("sync-pixels", pixels.Module, fx.Provide(New))
