package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command
func NewRootCmd(container *cli.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Version: container.Config.Version.VersionText(),
		Use:     "echoy",
		Short:   "Your intelligent CLI assistant",
		Long: `Echoy - Where your questions echo back with enhanced intelligence.
            
            A smart CLI assistant that transforms your queries into insightful 
            responses, creating a true dialogue between you and technology.`,
		RunE: func(cm *cobra.Command, args []string) error {
			themeManager := container.ThemeMgr
			themeManager.DisplayBanner(fmt.Sprintf("Welcome to %s", container.Config.Name), 40, "Your AI assistant for the CLI")
			fmt.Println("")
			themeManager.GetCurrentTheme().Warning().Println("Please run 'echoy init' to set up your assistant.")

			telemetryEvent.SendTelemetryEvent(
				cm.Context(),
				container.Config,
				"cmd.root.execute",
				telemetry.SeverityInfo,
				"Root command executed",
				nil,
			)

			return nil
		},
	}

	return rootCmd
}
