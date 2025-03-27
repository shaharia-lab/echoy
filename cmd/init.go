package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/spf13/cobra"
)

// NewInitCmd creates an interactive init command
func NewInitCmd(config *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Version: config.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			initializer := initPkg.NewInitializer(logger, config, themeManager)
			if err := initializer.Run(); err != nil {
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Initialization failed: %v", err))
				return err
			}

			themeManager.GetCurrentTheme().Info().Println("Run 'echoy help' to see the available commands.")
			return nil
		},
	}

	return cmd
}
