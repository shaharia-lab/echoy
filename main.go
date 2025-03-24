package main

import (
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/logger"
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

	// setup filesystem
	paths, err := filesystem.NewAppFilesystem(appCfg).EnsureAllPaths()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to ensure all application paths: %w", err))
		os.Exit(1)
	}

	// setup logger
	log, err := logger.NewLogger(logger.Config{
		FilePath: paths[filesystem.LogsFilePath],
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

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
		log.Sync()
		os.Exit(1)
	}

	log.Infof(fmt.Sprintf("%s exited successfully", appCfg.Name))
	log.Sync()
}
