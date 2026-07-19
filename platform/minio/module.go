package minio

import "go.uber.org/fx"

// Module provides the configured MinIO client from an injected Config.
var Module = fx.Module("minio", fx.Provide(New))
