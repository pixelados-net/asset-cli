package stats

import "go.uber.org/fx"

// Module provides the stats realm's service from the injected MinIO client.
var Module = fx.Module("stats", fx.Provide(NewService))
