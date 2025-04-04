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
func NewCmd(config *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager, initializer *Initializer) *cobra.Command {
	cmd := &cobra.Command{
		Version: config.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			telemetryEvent.SendTelemetryEvent(
				context.Background(),
				config,
				"cmd.init",
				telemetry.SeverityInfo, "Starting initialization",
				nil,
			)

			logger.Info("Starting initialization...")
			defer logger.Sync()

			if err := initializer.Run(); err != nil {
				logger.Error(fmt.Sprintf("Initialization failed: %v", err))
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Initialization failed: %v", err))
				return err
			}

			logger.Info("Initialization complete. You can now run 'echoy' to start using Echoy.")

			themeManager.GetCurrentTheme().Info().Println("\nRun 'echoy chat' to start an interactive chat session.")
			themeManager.GetCurrentTheme().Info().Println("Run 'echoy help' to see the available commands.")

			return nil
		},
	}

	return cmd
}
