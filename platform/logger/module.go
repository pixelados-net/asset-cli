package logger

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the process logger from an injected Config.
var Module = fx.Module("logger", fx.Provide(provideLogger))

func provideLogger(lifecycle fx.Lifecycle, config Config) *zap.Logger {
	log := New(config)
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		_ = log.Sync()
		return nil
	}})
	return log
}
