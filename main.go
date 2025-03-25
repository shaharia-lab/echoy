package main

import (
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"os"
)

var version = "0.0.1"
var commit = "none"
var date = "unknown"

func main() {
	// Initialize with configurable options
	if err := cli.InitWithOptions(cli.InitOptions{
		Version:  version,
		Commit:   commit,
		Date:     date,
		LogLevel: logger.DebugLevel,
		Theme:    theme.Professional,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error during initialization: %v\n", err)
		os.Exit(1)
	}

	appCfg := cli.GetAppConfig()
	log := cli.GetLogger()
	defer log.Sync()

	log.Infof(fmt.Sprintf("%s started", appCfg.Name))

	// setup commands
	rootCmd := cmd.NewRootCmd()
	rootCmd.AddCommand(
		cmd.NewInitCmd(),
		cmd.NewConfigCmd(appCfg, log),
		cmd.NewChatCmd(appCfg, log),
		cmd.NewUpdateCmd(),
	)

	// execute the command
	if err := rootCmd.Execute(); err != nil {
		cli.GetTheme().Error().Println(err)
		//log.Error(fmt.Sprintf("%v", err))
		log.Sync()
		os.Exit(1)
	}

	log.Infof(fmt.Sprintf("%s exited successfully", appCfg.Name))
	log.Sync()
}
