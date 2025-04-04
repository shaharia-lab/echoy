package main

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/cmd"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/cli"
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

	container, err := cli.NewContainer(cli.InitOptions{
		Version:  version,
		Commit:   commit,
		Date:     date,
		LogLevel: logger.InfoLevel,
		Theme:    theme.NewProfessionalTheme(),
	})
	if err != nil {
		telemetryEvent.SendTelemetryEvent(ctx, container.Config, "start.failed", telemetry.SeverityError, fmt.Sprintf("Error initializing container: %v", err), nil)
		fmt.Println("Error initializing container:", err)
		os.Exit(1)
	}

	telemetryEvent.SendTelemetryEvent(ctx, container.Config, "start", telemetry.SeverityInfo, "CLI starting", nil)

	// setup commands
	rootCmd := cmd.NewRootCmd(container)
	rootCmd.AddCommand(
		initializer.NewCmd(container.Config, container.Logger, container.ThemeMgr, container.Initializer),
		chat.NewChatCmd(container),
		cmd.NewUpdateCmd(container.Config, container.ThemeMgr),
	)

	// execute the command
	if err := rootCmd.Execute(); err != nil {
		telemetryEvent.SendTelemetryEvent(ctx, container.Config, "root.cmd.error", telemetry.SeverityError, "Error executing command", map[string]interface{}{"error": err})
		fmt.Println(err)
		os.Exit(1)
	}
}
