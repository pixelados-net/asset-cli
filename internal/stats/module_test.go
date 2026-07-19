package stats

import (
	"testing"

	"go.uber.org/fx"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

func TestModuleGraph(t *testing.T) {
	if err := fx.ValidateApp(bootstrap.Module, Module, fx.NopLogger); err != nil {
		t.Fatalf("ValidateApp() error = %v", err)
	}
}
