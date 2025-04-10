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
	"os"
	"text/tabwriter"
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

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

			fmt.Fprintln(w, "COMPONENT\tSTATUS\tDETAILS")

			provider := &UnixSocketProvider{
				SocketPath: socketPath,
				Timeout:    500 * time.Millisecond,
			}

			client := NewClient(provider, 500*time.Millisecond, 2*time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			isRunning, status := client.IsRunning(ctx)

			if isRunning {
				fmt.Fprintln(w, fmt.Sprintf("daemon\trunning\t-"))
				w.Flush()
				themeManager.GetCurrentTheme().Success().Println("\nDaemon is running correctly")
			} else {
				fmt.Fprintln(w, fmt.Sprintf("daemon\t%s\t-", status))
				w.Flush()
				themeManager.GetCurrentTheme().Warning().Println("\nDaemon is not running. Start it with 'echoy daemon start'")
			}

			return nil
		},
	}
	return cmd
}
