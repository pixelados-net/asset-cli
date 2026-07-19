package pets

import "go.uber.org/fx"

// Module provides the pets realm and Nitro's standard pet asset catalog.
var Module = fx.Module("pets", fx.Provide(NewService, NewClientCatalog))
