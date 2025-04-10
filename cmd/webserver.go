package cmd

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/daemon"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

// NewWebserverCmd creates a command to manage the webserver through the daemon
func NewWebserverCmd(container *cli.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webserver [start|stop]",
		Short: "Manage the Echoy web server",
		Long:  `Start or stop the Echoy web server through the daemon.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			defer container.Logger.Flush()

			subcommand := strings.ToLower(args[0])
			if subcommand != "start" && subcommand != "stop" {
				container.Logger.WithFields(map[string]interface{}{
					logger.ErrorKey: fmt.Errorf("invalid subcommand: %s", subcommand),
					"command":       "webserver",
					"subcommand":    subcommand,
				}).Error("invalid subcommand")

				return fmt.Errorf("invalid subcommand: %s (must be 'start' or 'stop')", subcommand)
			}

			provider := &daemon.UnixSocketProvider{
				SocketPath: container.SocketFilePath,
				Timeout:    500 * time.Millisecond,
			}
			client := daemon.NewClient(provider, 2*time.Second, 5*time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			response, err := client.Execute(ctx, "webserver", []string{subcommand})
			if err != nil {
				if isConnectionError(err) {
					msg := "Daemon is not running. Please start the daemon first with 'echoy start'"

					container.Logger.WithFields(map[string]interface{}{
						logger.ErrorKey: err,
						"command":       "webserver",
						"subcommand":    subcommand,
					}).Error("webserver command failed because the daemon is not running")

					container.ThemeMgr.GetCurrentTheme().Error().Println(msg)
					return fmt.Errorf("daemon is not running")
				}

				container.Logger.WithFields(map[string]interface{}{
					logger.ErrorKey: err,
					"command":       "webserver",
					"subcommand":    subcommand,
				}).Error("failed to execute webserver command")

				container.ThemeMgr.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to %s webserver: %v", subcommand, err))
				return fmt.Errorf("failed to %s webserver: %w", subcommand, err)
			}

			response = strings.TrimSpace(response)
			if strings.HasPrefix(response, "OK: ") {
				response = strings.TrimPrefix(response, "OK: ")
			}

			container.Logger.WithFields(map[string]interface{}{
				"command":    "webserver",
				"subcommand": subcommand,
				"response":   response,
			}).Info("Webserver command executed")

			container.ThemeMgr.GetCurrentTheme().Success().Println(response)
			return nil
		},
	}

	cmd.Example = "  echoy webserver start  # Start the web server\n" +
		"  echoy webserver stop   # Stop the web server"

	return cmd
}

// isConnectionError checks if the error is related to connection issues
func isConnectionError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such file or directory") ||
		strings.Contains(errStr, "socket operation on non-socket")
}
