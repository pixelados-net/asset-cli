// Package cli assembles the sync realm's Cobra command tree. It is the one place
// that sees both the sync core (internal/sync) and every emulator adapter
// (internal/sync/arcturus, internal/sync/pixels): the core package must not import
// adapters, or it would cycle back into itself, so branching on which adapter to
// compose lives here instead.
package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	syncrealm "github.com/pixelados-net/asset-cli/internal/sync"
	syncarcturus "github.com/pixelados-net/asset-cli/internal/sync/arcturus"
	syncpixels "github.com/pixelados-net/asset-cli/internal/sync/pixels"
	"github.com/pixelados-net/asset-cli/platform/bootstrap"
	"github.com/pixelados-net/asset-cli/platform/config"
)

// errMissingDefinitions is returned when check finds at least one client
// definition missing from the emulator; the detail was already logged.
var errMissingDefinitions = errors.New("client declares furniture definitions missing from the emulator")

// NewRealmCommand builds the sync realm's Cobra command tree.
func NewRealmCommand() *cobra.Command {
	realm := &cobra.Command{
		Use:   "sync",
		Short: "reconcile the client's catalogs with the configured emulator",
	}
	realm.AddCommand(newFurnitureCommand())
	return realm
}

func newFurnitureCommand() *cobra.Command {
	furniture := &cobra.Command{
		Use:   "furniture",
		Short: "sync furniture definitions between the client and the emulator",
	}
	furniture.AddCommand(newCheckCommand(), newApplyCommand())
	return furniture
}

// invokeWithEmulator resolves ASSET_CLI_EMULATOR_KIND once, up front, so exactly
// one of the Arcturus or Pixels backends is composed into the Fx graph — never
// both, and never neither.
func invokeWithEmulator(ctx context.Context, fn any) error {
	if err := config.LoadDotenv(); err != nil {
		return fmt.Errorf("load dotenv: %w", err)
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var emulatorModule fx.Option
	switch cfg.Emulator {
	case config.EmulatorArcturus:
		emulatorModule = syncarcturus.Module
	case config.EmulatorPixels:
		emulatorModule = syncpixels.Module
	default:
		return fmt.Errorf("unsupported emulator kind %q", cfg.Emulator)
	}

	return bootstrap.Invoke(ctx, fx.Options(syncrealm.Module, emulatorModule), fn)
}

func newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "diff the client's furniture catalog against the configured emulator",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			var incomplete bool
			err := invokeWithEmulator(ctx, func(service syncrealm.Service, log *zap.Logger) error {
				report, checkErr := service.Check(ctx)
				if checkErr != nil {
					log.Error("sync furniture check failed", zap.Error(checkErr))
					return checkErr
				}
				for _, classname := range report.Missing {
					log.Warn("client declares a definition missing from the emulator", zap.String("classname", classname))
				}
				for _, classname := range report.Orphaned {
					log.Info("emulator has a definition the client no longer declares", zap.String("classname", classname))
				}
				for _, change := range report.NameChanges {
					log.Info("naming differs between client and emulator",
						zap.String("classname", change.Classname),
						zap.String("clientName", change.ClientName),
						zap.String("emulatorName", change.EmulatorName),
					)
				}
				log.Info("sync furniture check summary",
					zap.Int("missing", len(report.Missing)),
					zap.Int("orphaned", len(report.Orphaned)),
					zap.Int("nameChanges", len(report.NameChanges)),
				)
				incomplete = !report.OK()
				return nil
			})
			if err != nil {
				return err
			}
			if incomplete {
				return errMissingDefinitions
			}
			return nil
		},
	}
}

func newApplyCommand() *cobra.Command {
	var confirm bool
	command := &cobra.Command{
		Use:   "apply",
		Short: "insert missing definitions and overwrite emulator naming to match the client",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return invokeWithEmulator(ctx, func(service syncrealm.Service, log *zap.Logger) error {
				if !confirm {
					report, checkErr := service.Check(ctx)
					if checkErr != nil {
						log.Error("sync furniture apply (dry run) failed", zap.Error(checkErr))
						return checkErr
					}
					log.Info("dry run: nothing written, pass --yes to write to the emulator",
						zap.Int("wouldInsert", len(report.Missing)),
						zap.Int("wouldUpdate", len(report.NameChanges)),
					)
					return nil
				}

				result, err := service.Apply(ctx)
				if err != nil {
					log.Error("sync furniture apply failed", zap.Error(err))
					return err
				}
				for _, classname := range result.Created {
					log.Info("inserted furniture definition", zap.String("classname", classname))
				}
				for _, classname := range result.Updated {
					log.Info("updated furniture definition naming", zap.String("classname", classname))
				}
				log.Info("sync furniture apply summary",
					zap.Int("inserted", len(result.Created)),
					zap.Int("updated", len(result.Updated)),
				)
				return nil
			})
		},
	}
	command.Flags().BoolVar(&confirm, "yes", false, "write to the emulator instead of a dry run")
	return command
}
