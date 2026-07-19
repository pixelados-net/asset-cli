package arcturus

import (
	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/arcturus"
)

// Module composes the Arcturus platform client and provides it as the sync
// realm's EmulatorCatalog.
var Module = fx.Module("sync-arcturus", arcturus.Module, fx.Provide(New))
