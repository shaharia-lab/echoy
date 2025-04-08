package main

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/daemon"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/initializer"
	"github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"os"
)

var version = "0.0.1"
var commit = "none"
var date = "unknown"

func main() {
	ctx := context.Background()

	cliContainer, err := cli.NewContainer(cli.InitOptions{
		Version:  version,
		Commit:   commit,
		Date:     date,
		LogLevel: logger.InfoLevel,
		Theme:    theme.NewProfessionalTheme(),
	})
	if err != nil {
		fmt.Println("Error initializing cliContainer:", err)
		os.Exit(1)
	}

	if cliContainer.ConfigFromFile.UsageTracking.Enabled {
		telemetryEvent.SendTelemetryEvent(ctx, cliContainer.Config, "start", telemetry.SeverityInfo, "CLI starting", nil)
	}

	// setup commands
	rootCmd := cmd.NewRootCmd(cliContainer)
	rootCmd.AddCommand(
		initializer.NewCmd(cliContainer.ConfigFromFile, cliContainer.Config, cliContainer.Logger, cliContainer.ThemeMgr, cliContainer.Initializer),
		chat.NewChatCmd(cliContainer),
		cmd.NewUpdateCmd(cliContainer.ConfigFromFile, cliContainer.Config, cliContainer.ThemeMgr),
		daemon.NewStartCmd(cliContainer.ConfigFromFile, cliContainer.Config, cliContainer.Logger, cliContainer.ThemeMgr, cliContainer.SocketFilePath, cliContainer.Paths[filesystem.CacheWebuiBuild], cliContainer.Logger),
		daemon.NewStopCmd(cliContainer.ConfigFromFile, cliContainer.Config, cliContainer.Logger, cliContainer.ThemeMgr, cliContainer.SocketFilePath),
		daemon.NewStatusCmd(cliContainer.ConfigFromFile, cliContainer.Config, cliContainer.Logger, cliContainer.ThemeMgr, cliContainer.SocketFilePath),
	)

	// execute the command
	if err := rootCmd.Execute(); err != nil {
		if cliContainer.ConfigFromFile.UsageTracking.Enabled {
			telemetryEvent.SendTelemetryEvent(ctx, cliContainer.Config, "root.cmd.error", telemetry.SeverityError, "Error executing command", map[string]interface{}{"error": err})
		}
		fmt.Println(err)
		os.Exit(1)
	}
}
