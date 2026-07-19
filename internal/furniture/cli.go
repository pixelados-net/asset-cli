package furniture

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

// errMissingBundles is returned when check finds at least one catalog entry with no
// matching bundle file; the detail was already logged per classname.
var errMissingBundles = errors.New("furniture catalog references bundles missing from the bucket")

// NewRealmCommand builds the furniture realm's Cobra command tree.
func NewRealmCommand() *cobra.Command {
	realm := &cobra.Command{
		Use:   "furniture",
		Short: "cross-check furniture bundles against the furniture catalog",
	}
	realm.AddCommand(newCheckCommand())
	return realm
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "cross-check furniture/bundles against FurnitureData.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			var incomplete bool
			err := bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				report, checkErr := service.Check(ctx)
				if checkErr != nil {
					log.Error("furniture check failed", zap.Error(checkErr))
					return checkErr
				}
				for _, classname := range report.Orphaned {
					log.Warn("orphaned furniture bundle", zap.String("classname", classname))
				}
				for _, classname := range report.Missing {
					log.Warn("furniture catalog entry missing its bundle", zap.String("classname", classname))
				}
				log.Info("furniture check summary",
					zap.Int("matched", report.Matched),
					zap.Int("orphaned", len(report.Orphaned)),
					zap.Int("missing", len(report.Missing)),
				)
				incomplete = !report.OK()
				return nil
			})
			if err != nil {
				return err
			}
			if incomplete {
				return errMissingBundles
			}
			return nil
		},
	}
}
