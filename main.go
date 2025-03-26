package main

import (
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/logger"
	"os"
)

var version = "0.0.1"
var commit = "none"
var date = "unknown"

func main() {
	container, err := cli.NewContainer(cli.InitOptions{
		Version:  "1.0.0",
		Commit:   "abc123",
		Date:     "2023-11-01",
		LogLevel: logger.InfoLevel,
	})

	if err != nil {
		fmt.Printf("failed to initialize container: %v\n", err)
		os.Exit(1)
	}

	log := container.Logger
	appCfg := container.Config

	defer container.Logger.Sync()

	log.Infof(fmt.Sprintf("%s started", appCfg.Name))

	// setup commands
	rootCmd := cmd.NewRootCmd(container)
	rootCmd.AddCommand(
		cmd.NewInitCmd(container),
		cmd.NewConfigCmd(appCfg, log),
		chat.NewChatCmd(container),
		cmd.NewUpdateCmd(container),
	)

	// execute the command
	if err := rootCmd.Execute(); err != nil {
		container.ThemeMgr.GetCurrentTheme().Error().Println(err)
		//log.Error(fmt.Sprintf("%v", err))
		log.Sync()
		os.Exit(1)
	}

	log.Infof(fmt.Sprintf("%s exited successfully", appCfg.Name))
	log.Sync()
}
