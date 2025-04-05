package daemon

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewStatusCmd creates a command to check the daemon status
func NewStatusCmd(config config.Config, appConfig *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager, socketPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check the status of the Echoy daemon",
		Long:  `Checks if the Echoy daemon is currently running.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(),
					appConfig,
					"daemon.status",
					telemetry.SeverityInfo, "Checking daemon status",
					nil,
				)
			}

			logger.Info("Checking daemon status...")
			defer logger.Sync()

			// Try to connect to the daemon socket
			conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
			if err != nil {
				logger.Info("Daemon is not running")
				themeManager.GetCurrentTheme().Info().Println("Daemon status: Not running")
				return nil
			}
			defer conn.Close()

			// Send a PING command to verify it's responsive
			_, err = conn.Write([]byte("PING\n"))
			if err != nil {
				logger.Warn(fmt.Sprintf("Daemon socket exists but may not be responsive: %v", err))
				themeManager.GetCurrentTheme().Warning().Println("Daemon status: Socket exists but daemon is not responsive")
				return nil
			}

			// Read response
			buffer := make([]byte, 128)
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			n, err := conn.Read(buffer)
			if err != nil || n == 0 {
				logger.Warn("Daemon did not respond to ping")
				themeManager.GetCurrentTheme().Warning().Println("Daemon status: Socket exists but daemon is not responding to commands")
				return nil
			}

			response := string(buffer[:n])
			if strings.TrimSpace(response) == "PONG" {
				logger.Info("Daemon is running and responsive")
				themeManager.GetCurrentTheme().Success().Println("Daemon status: Running")
				return nil
			}

			logger.Warn(fmt.Sprintf("Daemon sent unexpected response: %s", response))
			themeManager.GetCurrentTheme().Warning().Println("Daemon status: Running but returned unexpected response")
			return nil
		},
	}
	return cmd
}

// ResolveSocketPath converts template paths into actual file paths
// It handles environment variables and user directories
func ResolveSocketPath(templatePath string) string {
	socketPath := templatePath

	// Handle $HOME variable replacement, matching NewDaemon's logic
	if strings.Contains(socketPath, "$HOME") {
		home, err := os.UserHomeDir()
		if err == nil {
			socketPath = strings.Replace(socketPath, "$HOME", home, 1)
		}
	}

	// Also handle XDG_RUNTIME_DIR if present in the path
	if strings.Contains(socketPath, "$XDG_RUNTIME_DIR") {
		runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
		if runtimeDir == "" {
			// Fallback if XDG_RUNTIME_DIR is not set
			home, err := os.UserHomeDir()
			if err == nil {
				runtimeDir = filepath.Join(home, ".echoy")
			} else {
				// Ultimate fallback
				runtimeDir = "/tmp"
			}
		}
		socketPath = strings.Replace(socketPath, "$XDG_RUNTIME_DIR", runtimeDir, 1)
	}

	return socketPath
}

// resolveSocketPath is an alias for backward compatibility
func resolveSocketPath(templatePath string) string {
	return ResolveSocketPath(templatePath)
}
