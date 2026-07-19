// Package cli contains the asset-cli Cobra command tree.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/pixelados-net/asset-cli/internal/clothing"
	"github.com/pixelados-net/asset-cli/internal/effects"
	"github.com/pixelados-net/asset-cli/internal/furniture"
	"github.com/pixelados-net/asset-cli/internal/pets"
	"github.com/pixelados-net/asset-cli/internal/stats"
	"github.com/pixelados-net/asset-cli/internal/structure"
	synccli "github.com/pixelados-net/asset-cli/internal/sync/cli"
)

// NewRootCommand builds the asset-cli root command with every realm's command
// tree attached. This is the one place that assembles realms into the CLI transport;
// each realm otherwise stays unaware of any other realm or of Cobra internals beyond
// its own command tree.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "asset-cli",
		Short:         "asset-cli normalizes and manages Habbo asset storage",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCommand(version))
	root.AddCommand(structure.NewRealmCommand())
	root.AddCommand(stats.NewRealmCommand())
	root.AddCommand(furniture.NewRealmCommand())
	root.AddCommand(clothing.NewRealmCommand())
	root.AddCommand(effects.NewRealmCommand())
	root.AddCommand(pets.NewRealmCommand())
	root.AddCommand(synccli.NewRealmCommand())
	return root
}
