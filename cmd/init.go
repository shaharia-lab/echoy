package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/spf13/cobra"
)

// NewInitCmd creates an interactive init command
func NewInitCmd(appCfg *config.AppConfig, log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			initializer := initPkg.NewInitializer(log, appCfg)
			if err := initializer.Run(); err != nil {
				log.Error(fmt.Sprintf("Initialization failed: %v", err))
				return err
			}

			color.Green("Echoy has been successfully configured!")
			return nil
		},
	}

	return cmd
}
