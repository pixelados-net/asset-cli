package bootstrap

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/config"
)

const shutdownTimeout = 15 * time.Second

// Invoke starts the Fx graph composed of the platform Module plus a realm's own
// module, resolves fn's parameters, runs fn once, and stops the graph. Realm
// commands use this instead of constructing MinIO clients or loggers directly,
// keeping their logic reusable from any transport.
func Invoke(ctx context.Context, realm fx.Option, fn any) error {
	if err := config.LoadDotenv(); err != nil {
		return fmt.Errorf("load dotenv: %w", err)
	}
	fxApp := fx.New(Module, realm, fx.Invoke(fn), fx.NopLogger)
	startErr := fxApp.Start(ctx)
	stopContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	stopErr := fxApp.Stop(stopContext)
	if startErr != nil {
		return fmt.Errorf("start application: %w", startErr)
	}
	return stopErr
}
