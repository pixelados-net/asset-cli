package structure

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

// errIncompleteStructure is returned when check finds at least one missing path,
// so the process exits non-zero; the detail was already logged per path.
var errIncompleteStructure = errors.New("bucket layout is incomplete")

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
			var incomplete bool
			err := bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				report, checkErr := service.Check(ctx)
				if checkErr != nil {
					log.Error("structure check failed", zap.Error(checkErr))
					return checkErr
				}
				for _, path := range report.Present {
					log.Info("structure path present", zap.String("path", path))
				}
				for _, path := range report.Missing {
					log.Warn("structure path missing", zap.String("path", path))
				}
				incomplete = !report.OK()
				return nil
			})
			if err != nil {
				return err
			}
			if incomplete {
				return errIncompleteStructure
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
			return bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				created, createErr := service.Create(ctx)
				if createErr != nil {
					log.Error("structure create failed", zap.Error(createErr))
					return createErr
				}
				if len(created) == 0 {
					log.Info("bucket layout already complete")
					return nil
				}
				for _, path := range created {
					log.Info("created structure path", zap.String("path", path))
				}
				return nil
			})
		},
	}
}
