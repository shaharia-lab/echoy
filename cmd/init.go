package cmd

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/spf13/cobra"
)

// NewInitCmd creates an interactive init command
func NewInitCmd(c *cli.Container) *cobra.Command {
	cmd := &cobra.Command{
		Version: c.Config.Version.VersionText(),
		Use:     "init",
		Short:   "Initialize the Echoy with a guided setup",
		Long:    `Start an interactive wizard to configure Echoy with a series of questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := c.Logger
			initializer := initPkg.NewInitializer(log, c.Config, c.ThemeMgr)
			if err := initializer.Run(); err != nil {
				c.ThemeMgr.GetCurrentTheme().Error().Println(fmt.Sprintf("Initialization failed: %v", err))
				return err
			}

			c.ThemeMgr.GetCurrentTheme().Info().Println("Run 'echoy help' to see the available commands.")
			return nil
		},
	}

	return cmd
}
