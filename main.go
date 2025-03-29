package main

import (
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/initializer"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"os"
)

var version = "0.0.1"
var commit = "none"
var date = "unknown"

func main() {
	container, err := cli.NewContainer(cli.InitOptions{
		Version:  version,
		Commit:   commit,
		Date:     date,
		LogLevel: logger.InfoLevel,
		Theme:    theme.NewProfessionalTheme(),
	})

	if err != nil {
		fmt.Println("Error initializing container:", err)
		os.Exit(1)
	}

	// setup commands
	rootCmd := cmd.NewRootCmd(container)
	rootCmd.AddCommand(
		initializer.NewCmd(container.Config, container.Logger, container.ThemeMgr, container.Initializer),
		chat.NewChatCmd(container),
		cmd.NewUpdateCmd(container.Config, container.ThemeMgr),
	)

	// execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
