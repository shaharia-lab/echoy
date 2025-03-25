package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/spf13/cobra"
	"os"
)

// NewConfigCmd creates a config command
func NewConfigCmd(appCfg *config.AppConfig, log *logger.Logger) *cobra.Command {
	cfgCmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "config",
		Short:   "Manage Echoy configuration",
		Long:    `Commands to manage and view your Echoy configuration.`,
	}

	cfgCmd.AddCommand(NewConfigPreviewCmd(log))
	return cfgCmd
}

// NewConfigPreviewCmd creates a command to preview the config file
func NewConfigPreviewCmd(log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Preview the current configuration file",
		Long:  `Display the content of your Echoy configuration file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log.WithFields(map[string]interface{}{"test": "test2"}).Debug("Previewing configuration file")

			log.Debug("Previewing configuration file")

			configPath := config.GetConfigPath()
			configData, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			color.New(color.FgHiCyan, color.Bold).Println("\n📄 Configuration File")
			color.New(color.FgHiWhite).Printf("Located at: %s\n\n", configPath)

			// Print the YAML content
			fmt.Println(string(configData))

			return nil
		},
	}

	return cmd
}
