package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command
func NewRootCmd(c *cli.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: c.Config.Version.VersionText(),
		Use:     "echoy",
		Short:   "Your intelligent CLI assistant",
		Long: `Echoy - Where your questions echo back with enhanced intelligence.
            
            A smart CLI assistant that transforms your queries into insightful 
            responses, creating a true dialogue between you and technology.`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			theme.DisplayBanner(c.Config)

			// Skip check for init and help commands
			if cmd.Name() == "init" || cmd.Name() == "help" {
				return
			}

			if !config.ConfigExists() {
				c.ThemeMgr.GetCurrentTheme().Error().Println("Configuration not found. Please run 'echoy init' to set up.")
				c.Logger.Error("Configuration not found. Please run 'echoy init' to set up.")
				c.Logger.Sync()
			}

			return
		},
		RunE: func(cm *cobra.Command, args []string) error {
			if config.ConfigExists() {
				c.ThemeMgr.GetCurrentTheme().Info().Println(fmt.Sprintf("%s is configured and ready to use!", c.Config.Name))
				return nil
			}

			return nil
		},
	}

	return rootCmd
}
