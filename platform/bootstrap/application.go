package bootstrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pixelados-net/asset-cli/platform/minio"
)

const shutdownTimeout = 15 * time.Second

// Application owns the Fx graph backing a single CLI command invocation.
type Application struct {
	app       *fx.App
	Logger    *zap.Logger
	Storage   *minio.Client
	closeOnce sync.Once
}

// New builds and starts the Fx application graph.
func New(ctx context.Context) (*Application, error) {
	var log *zap.Logger
	var storage *minio.Client
	fxApp := fx.New(Module, fx.Populate(&log, &storage), fx.NopLogger)
	if err := fxApp.Start(ctx); err != nil {
		stopContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = fxApp.Stop(stopContext)
		return nil, fmt.Errorf("start application: %w", err)
	}
	return &Application{app: fxApp, Logger: log, Storage: storage}, nil
}

// Close stops the Fx graph once in reverse dependency order.
func (application *Application) Close() {
	application.closeOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = application.app.Stop(ctx)
	})
}
