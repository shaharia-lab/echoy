package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/spf13/cobra"
)

// NewInitCmd creates an interactive init command
func NewInitCmd() *cobra.Command {
	appCfg := cli.GetAppConfig()
	cliTheme := cli.GetTheme()

	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := cli.GetLogger()
			initializer := initPkg.NewInitializer(log, appCfg, cliTheme)
			if err := initializer.Run(); err != nil {
				cliTheme.Error().Println(fmt.Sprintf("Initialization failed: %v", err))
				return err
			}

			cliTheme.Info().Println("Run 'echoy help' to see the available commands.")
			return nil
		},
	}

	return cmd
}
