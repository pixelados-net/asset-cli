package pets

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

var errMissingBundles = errors.New("nitro pet types reference bundles missing from the bucket")

// NewRealmCommand builds the pets realm's Cobra command tree.
func NewRealmCommand() *cobra.Command {
	realm := &cobra.Command{Use: "pets", Short: "cross-check pet bundles against Nitro pet types"}
	realm.AddCommand(newCheckCommand())
	return realm
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use: "check", Short: "cross-check pets bundles against Nitro pet types",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			var incomplete bool
			err := bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				report, checkErr := service.Check(ctx)
				if checkErr != nil {
					log.Error("pets check failed", zap.Error(checkErr))
					return checkErr
				}
				for _, name := range report.Orphaned {
					log.Warn("orphaned pet bundle", zap.String("asset", name))
				}
				for _, name := range report.Missing {
					log.Warn("pet type missing its bundle", zap.String("asset", name))
				}
				log.Info("pets check summary", zap.Int("matched", report.Matched),
					zap.Int("orphaned", len(report.Orphaned)), zap.Int("missing", len(report.Missing)))
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
