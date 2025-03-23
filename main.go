package main

import (
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/config"
	"os"
)

var version = "0.0.1"
var commit = "none"
var date = "unknown"

func main() {
	appCfg := config.NewDefaultConfig(
		config.WithVersion(config.Version{
			Version: version,
			Commit:  commit,
			Date:    date,
		}),
	)

	rootCmd := cmd.NewRootCmd(appCfg)
	rootCmd.AddCommand(
		cmd.NewInitCmd(appCfg),
		cmd.NewConfigCmd(appCfg),
		cmd.NewChatCmd(appCfg),
		cmd.NewUpdateCmd(appCfg),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
