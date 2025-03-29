package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/initializer"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/spf13/cobra"
)

// NewInitCmd creates an interactive init command
func NewInitCmd(config *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager, initializer *initializer.Initializer) *cobra.Command {
	cmd := &cobra.Command{
		Version: config.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Starting initialization...")

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
