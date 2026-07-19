package structure

import "go.uber.org/fx"

// Module provides the structure realm's service from the injected MinIO client.
var Module = fx.Module("structure", fx.Provide(NewService))
