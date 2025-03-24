package main

import (
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/cli"
	"os"
)

var version = "0.0.1"
var commit = "none"
var date = "unknown"

func main() {
	// Initialize with build information
	if err := cli.InitWithOptions(cli.InitOptions{
		Version: version,
		Commit:  commit,
		Date:    date,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Error during initialization: %v\n", err)
		os.Exit(1)
	}

	appCfg := cli.GetAppConfig()
	log := cli.GetLogger()
	defer log.Sync()

	log.Infof(fmt.Sprintf("%s started", appCfg.Name))

	// setup commands
	rootCmd := cmd.NewRootCmd(appCfg, log)
	rootCmd.AddCommand(
		cmd.NewInitCmd(appCfg, log),
		cmd.NewConfigCmd(appCfg, log),
		cmd.NewChatCmd(appCfg, log),
		cmd.NewUpdateCmd(appCfg, log),
	)

	// execute the command
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		log.Error(fmt.Sprintf("%s exited with error: %v", appCfg.Name, err))
		log.Sync()
		os.Exit(1)
	}

	log.Infof(fmt.Sprintf("%s exited successfully", appCfg.Name))
	log.Sync()
}
