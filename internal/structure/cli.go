package structure

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

// NewRealmCommand builds the structure realm's Cobra command tree.
func NewRealmCommand() *cobra.Command {
	realm := &cobra.Command{
		Use:   "structure",
		Short: "manage the asset-cli bucket layout",
	}
	realm.AddCommand(newCheckCommand(), newCreateCommand())
	return realm
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "verify every expected bucket path exists",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			var report Report
			err := bootstrap.Invoke(ctx, Module, func(service Service) error {
				checked, checkErr := service.Check(ctx)
				report = checked
				return checkErr
			})
			if err != nil {
				return err
			}
			writeReport(cmd, report)
			if !report.OK() {
				return fmt.Errorf("%d expected path(s) missing", len(report.Missing))
			}
			return nil
		},
	}
}

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "create every missing expected bucket path",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			var created []string
			err := bootstrap.Invoke(ctx, Module, func(service Service) error {
				paths, createErr := service.Create(ctx)
				created = paths
				return createErr
			})
			if err != nil {
				return err
			}
			if len(created) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "nothing to create, bucket layout is complete")
				return err
			}
			for _, path := range created {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", path); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func writeReport(cmd *cobra.Command, report Report) {
	for _, path := range report.Present {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ok      %s\n", path)
	}
	for _, path := range report.Missing {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "missing %s\n", path)
	}
}
