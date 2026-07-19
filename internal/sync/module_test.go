// Package sync_test validates the sync realm's Fx graph with each emulator
// adapter. It is an external test package (not "package sync") specifically so
// it can import both internal/sync and internal/sync/{arcturus,pixels} without
// recreating the cycle those adapters avoid by not importing internal/sync/cli.
package sync_test

import (
	"testing"

	"go.uber.org/fx"

	syncrealm "github.com/pixelados-net/asset-cli/internal/sync"
	syncarcturus "github.com/pixelados-net/asset-cli/internal/sync/arcturus"
	syncpixels "github.com/pixelados-net/asset-cli/internal/sync/pixels"
	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

func TestModuleGraphWithArcturus(t *testing.T) {
	if err := fx.ValidateApp(bootstrap.Module, syncrealm.Module, syncarcturus.Module, fx.NopLogger); err != nil {
		t.Fatalf("ValidateApp() error = %v", err)
	}
}

func TestModuleGraphWithPixels(t *testing.T) {
	if err := fx.ValidateApp(bootstrap.Module, syncrealm.Module, syncpixels.Module, fx.NopLogger); err != nil {
		t.Fatalf("ValidateApp() error = %v", err)
	}
}
