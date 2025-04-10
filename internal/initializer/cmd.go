package initializer

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewCmd creates an interactive init command
func NewCmd(config config.Config, appConfig *config.AppConfig, logger logger.Logger, themeManager *theme.Manager, initializer *Initializer) *cobra.Command {
	cmd := &cobra.Command{
		Version: appConfig.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(),
					appConfig,
					"cmd.init",
					telemetry.SeverityInfo, "Starting initialization",
					nil,
				)
			}

			logger.Info("Starting initialization...", nil)
			defer logger.Sync()

			if err := initializer.Run(); err != nil {
				logger.Errorf("Initialization failed: %v", err)
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Initialization failed: %v", err))
				return err
			}

			logger.Info("Initialization complete. You can now run 'echoy' to start using Echoy.", nil)

			themeManager.GetCurrentTheme().Info().Println("\nRun 'echoy chat' to start an interactive chat session.")
			themeManager.GetCurrentTheme().Info().Println("Run 'echoy help' to see the available commands.")

			return nil
		},
	}

	return cmd
}
