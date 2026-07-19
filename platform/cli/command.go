// Package cli contains the asset-cli Cobra command tree.
package cli

import "github.com/spf13/cobra"

// NewRootCommand builds the asset-cli root command with every subcommand attached.
func NewRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "asset-cli",
		Short:         "asset-cli manages Habbo asset storage",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newVersionCommand(version))
	return root
}
