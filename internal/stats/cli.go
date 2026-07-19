package stats

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/pixelados-net/asset-cli/platform/bootstrap"
)

// NewRealmCommand builds the stats realm's Cobra command tree.
func NewRealmCommand() *cobra.Command {
	realm := &cobra.Command{
		Use:   "stats",
		Short: "report content counts for the asset-cli bucket",
	}
	realm.AddCommand(newNitroCommand(), newOrphanCommand())
	return realm
}

func newNitroCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "nitro",
		Short: "count .nitro bundles per content category",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				counts, err := service.Nitro(ctx)
				if err != nil {
					log.Error("stats nitro failed", zap.Error(err))
					return err
				}
				total := 0
				for _, count := range counts {
					log.Info("nitro bundle count", zap.String("category", count.Name), zap.Int("total", count.Total))
					total += count.Total
				}
				log.Info("nitro bundle total", zap.Int("total", total))
				return nil
			})
		},
	}
}

func newOrphanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "orphan",
		Short: "report orphaned and missing bundles per content category",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return bootstrap.Invoke(ctx, Module, func(service Service, log *zap.Logger) error {
				reports, err := service.Orphans(ctx)
				if err != nil {
					log.Error("stats orphan failed", zap.Error(err))
					return err
				}
				for _, report := range reports {
					log.Info("orphan report",
						zap.String("category", report.Category),
						zap.Int("matched", report.Matched),
						zap.Int("orphaned", report.Orphaned),
						zap.Int("missing", report.Missing),
					)
				}
				return nil
			})
		},
	}
}
