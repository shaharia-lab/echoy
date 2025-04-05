package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewStopCmd creates a command to stop the running daemon
func NewStopCmd(config config.Config, appConfig *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager) *cobra.Command {
	var socketPath string

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running Echoy daemon",
		Long:  `Sends a stop command to the running Echoy daemon to shut it down gracefully.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(),
					appConfig,
					"daemon.stop",
					telemetry.SeverityInfo, "Stopping daemon",
					nil,
				)
			}

			logger.Info("Stopping daemon...")
			defer logger.Sync()

			// If no custom socket path provided, use the default
			if socketPath == "" {
				socketPath = resolveSocketPath(DefaultSocketPath)
			}

			// Connect to the daemon socket
			conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to connect to daemon: %v", err)
				logger.Error(errMsg)
				themeManager.GetCurrentTheme().Error().Println(errMsg)
				return fmt.Errorf(errMsg)
			}
			defer conn.Close()

			// Send the stop command
			_, err = conn.Write([]byte("STOP\n"))
			if err != nil {
				errMsg := fmt.Sprintf("Failed to send stop command: %v", err)
				logger.Error(errMsg)
				themeManager.GetCurrentTheme().Error().Println(errMsg)
				return fmt.Errorf(errMsg)
			}

			// Optional: Wait for confirmation or timeout
			// This simple version just waits for the connection to close
			buffer := make([]byte, 128)
			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, err = conn.Read(buffer)

			themeManager.GetCurrentTheme().Success().Println("Daemon stopped successfully")
			logger.Info("Daemon stopped successfully")
			return nil
		},
	}

	// Add command-specific flags
	cmd.Flags().StringVar(&socketPath, "socket", "", fmt.Sprintf("Custom socket path (default: %s)", DefaultSocketPath))

	return cmd
}
