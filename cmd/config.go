package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/spf13/cobra"
	"os"
)

// NewConfigCmd creates a config command
func NewConfigCmd(appCfg *config.AppConfig) *cobra.Command {
	cfgCmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "config",
		Short:   "Manage Echoy configuration",
		Long:    `Commands to manage and view your Echoy configuration.`,
	}

	cfgCmd.AddCommand(NewConfigPreviewCmd())
	return cfgCmd
}

// NewConfigPreviewCmd creates a command to preview the config file
func NewConfigPreviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview the current configuration file",
		Long:  `Display the content of your Echoy configuration file.`,
		Run: func(cmd *cobra.Command, args []string) {
			configPath := config.GetConfigPath()
			configData, err := os.ReadFile(configPath)
			if err != nil {
				color.Red("Error reading config file: %v", err)
				os.Exit(1)
			}

			color.New(color.FgHiCyan, color.Bold).Println("\n📄 Configuration File")
			color.New(color.FgHiWhite).Printf("Located at: %s\n\n", configPath)

			// Print the YAML content
			fmt.Println(string(configData))
		},
	}

	return cmd
}
