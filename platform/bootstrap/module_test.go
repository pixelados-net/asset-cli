package bootstrap

import (
	"testing"

	"go.uber.org/fx"
)

func TestModuleGraph(t *testing.T) {
	if err := fx.ValidateApp(Module, fx.NopLogger); err != nil {
		t.Fatalf("ValidateApp() error = %v", err)
	}
}
