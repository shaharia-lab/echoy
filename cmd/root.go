package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command
func NewRootCmd(appCfg *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "echoy",
		Short:   "Your intelligent CLI assistant",
		Long: `Echoy - Where your questions echo back with enhanced intelligence.
            
            A smart CLI assistant that transforms your queries into insightful 
            responses, creating a true dialogue between you and technology.`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			defer logger.Sync()

			themeManager.DisplayBanner(fmt.Sprintf("Welcome to %s", appCfg.Name), 40, "Your AI assistant for the CLI")

			if cmd.Name() == "init" || cmd.Name() == "help" {
				return
			}

			if !config.ConfigExists() {
				themeManager.GetCurrentTheme().Error().Println("Configuration not found. Please run 'echoy init' to set up.")
				logger.Error("Configuration not found. Please run 'echoy init' to set up.")
			}

			return
		},
		RunE: func(cm *cobra.Command, args []string) error {
			if config.ConfigExists() {
				themeManager.GetCurrentTheme().Info().Println(fmt.Sprintf("%s is configured and ready to use!", appCfg.Name))
				return nil
			}

			return nil
		},
	}

	return rootCmd
}
