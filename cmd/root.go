package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/spf13/cobra"
	"os"
)

// NewRootCmd creates and returns the root command
func NewRootCmd(appCfg *config.AppConfig, log *logger.Logger) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "echoy",
		Short:   "Your intelligent CLI assistant",
		Long: `Echoy - Where your questions echo back with enhanced intelligence.
            
            A smart CLI assistant that transforms your queries into insightful 
            responses, creating a true dialogue between you and technology.`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip check for init and help commands
			if cmd.Name() == "init" || cmd.Name() == "help" {
				return
			}

			if !config.ConfigExists() {
				color.Yellow("Configuration not found. Please run 'echoy init' to set up.")
				os.Exit(1)
			}
		},
		Run: func(c *cobra.Command, args []string) {
			PrintColorfulBanner()
			if config.ConfigExists() {
				color.Green("Echoy is configured and ready to use!")
				fmt.Println("\nType 'echoy help' to see available commands.")
			} else {
				color.Yellow("Echoy needs to be configured.")
				fmt.Println("\nType 'echoy init' to start the configuration wizard!")
			}
		},
	}

	return rootCmd
}
