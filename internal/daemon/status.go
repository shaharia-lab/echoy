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

			conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
			if err != nil {
				logger.Info("Daemon is not running")
				themeManager.GetCurrentTheme().Info().Println("Daemon status: Not running")
				return nil
			}
			defer conn.Close()

			_, err = conn.Write([]byte("PING\n"))
			if err != nil {
				logger.Warn(fmt.Sprintf("Daemon socket exists but may not be responsive: %v", err))
				themeManager.GetCurrentTheme().Warning().Println("Daemon status: Socket exists but daemon is not responsive")
				return nil
			}

			// Read response
			buffer := make([]byte, 128)
			err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				return err
			}
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
