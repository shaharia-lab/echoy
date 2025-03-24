package cmd

import (
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/spf13/cobra"
	"os"
)

// NewInitCmd creates an interactive init command
func NewInitCmd(appCfg *config.AppConfig, log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		Run: func(cmd *cobra.Command, args []string) {
			initializer := initPkg.NewInitializer()
			if err := initializer.Run(); err != nil {
				color.Red("Initialization failed: %v", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
