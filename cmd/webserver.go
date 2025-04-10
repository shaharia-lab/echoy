package cmd

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/daemon"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/spf13/cobra"
	"log/slog"
	"strings"
	"time"
)

// NewWebserverCmd creates a command to manage the webserver through the daemon
func NewWebserverCmd(socketPath string, themeManager *theme.Manager, sLogger *slog.Logger, serverLogger logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webserver [start|stop]",
		Short: "Manage the Echoy web server",
		Long:  `Start or stop the Echoy web server through the daemon.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			subcommand := strings.ToLower(args[0])
			if subcommand != "start" && subcommand != "stop" {
				return fmt.Errorf("invalid subcommand: %s (must be 'start' or 'stop')", subcommand)
			}

			provider := &daemon.UnixSocketProvider{
				SocketPath: socketPath,
				Timeout:    500 * time.Millisecond,
			}
			client := daemon.NewClient(provider, 2*time.Second, 5*time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			response, err := client.Execute(ctx, "webserver", []string{subcommand})
			if err != nil {
				if isConnectionError(err) {
					msg := "Daemon is not running. Please start the daemon first with 'echoy start'"
					sLogger.Error(msg, "error", err)
					themeManager.GetCurrentTheme().Error().Println(msg)
					return fmt.Errorf("daemon is not running")
				}

				sLogger.Error("Failed to execute webserver command", "subcommand", subcommand, "error", err)
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to %s webserver: %v", subcommand, err))
				return fmt.Errorf("failed to %s webserver: %w", subcommand, err)
			}

			response = strings.TrimSpace(response)
			if strings.HasPrefix(response, "OK: ") {
				response = strings.TrimPrefix(response, "OK: ")
			}

			sLogger.Info(fmt.Sprintf("Webserver %s command executed", subcommand), "response", response)
			themeManager.GetCurrentTheme().Success().Println(response)
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
