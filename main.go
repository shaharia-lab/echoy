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

	fs := filesystem.NewAppFilesystem(appCfg)
	paths, err := fs.EnsureAllPaths()
	if err != nil {
		fmt.Println(fmt.Errorf("failed to ensure all application paths: %w", err))
		os.Exit(1)
	}

	log, err := logger.NewLogger(logger.Config{
		LogLevel:  logger.DebugLevel,
		FilePath:  paths[filesystem.LogsFilePath],
		MaxSizeMB: 1,
	})
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Add some test logs
	log.Info("Logger initialized successfully")
	log.WithFields(map[string]interface{}{"command": "config preview"}).Debug("Running command")

	// Force flush logs before any potential early exit
	log.Sync() // Explicitly flush logs right after writing them

	rootCmd := cmd.NewRootCmd(appCfg, log)
	rootCmd.AddCommand(
		cmd.NewInitCmd(appCfg, log),
		cmd.NewConfigCmd(appCfg, log),
		cmd.NewChatCmd(appCfg, log),
		cmd.NewUpdateCmd(appCfg, log),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		log.Sync()
		os.Exit(1)
	}

	log.Info("Command execution completed")
	log.Sync()
}
