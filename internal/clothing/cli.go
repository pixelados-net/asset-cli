package clothing

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

var errMissingBundles = errors.New("figure map references clothing bundles missing from the bucket")

// NewRealmCommand builds the clothing realm's Cobra command tree.
func NewRealmCommand() *cobra.Command {
	realm := &cobra.Command{Use: "clothing", Short: "cross-check clothing bundles against FigureMap.json"}
	realm.AddCommand(newCheckCommand())
	return realm
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use: "check", Short: "cross-check avatar/clothing against FigureMap.json",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			var incomplete bool
			err := bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				report, checkErr := service.Check(ctx)
				if checkErr != nil {
					log.Error("clothing check failed", zap.Error(checkErr))
					return checkErr
				}
				for _, name := range report.Orphaned {
					log.Warn("orphaned clothing bundle", zap.String("library", name))
				}
				for _, name := range report.Missing {
					log.Warn("clothing library missing its bundle", zap.String("library", name))
				}
				log.Info("clothing check summary", zap.Int("matched", report.Matched),
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
