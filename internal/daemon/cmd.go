package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewRunCmd creates a command to run the daemon
func NewRunCmd(config config.Config, appConfig *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager, socketPath string) *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Start the Echoy daemon",
		Long:  `Starts the Echoy daemon that processes background tasks and client requests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(),
					appConfig,
					"daemon.run",
					telemetry.SeverityInfo, "Starting daemon",
					nil,
				)
			}

			logger.Info("Starting daemon...")
			defer logger.Sync()

			// Create daemon instance
			daemon := NewDaemon(socketPath)

			// Create a context that can be canceled
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set up signal handling
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigChan
				logger.Info(fmt.Sprintf("Received signal %s, shutting down...", sig))
				cancel()
			}()

			// Start the daemon
			if err := daemon.Start(); err != nil {
				logger.Error(fmt.Sprintf("Failed to start daemon: %v", err))
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon: %v", err))
				return err
			}

			themeManager.GetCurrentTheme().Success().Printf("Daemon started and listening on %s\n", daemon.SocketPath)
			logger.Info(fmt.Sprintf("Daemon started and listening on %s", daemon.SocketPath))

			if !foreground {
				logger.Info("Daemon running in background mode")
				return nil
			}

			// Wait for context cancellation (signal or error)
			<-ctx.Done()
			logger.Info("Stopping daemon...")
			daemon.Stop()
			logger.Info("Daemon stopped")

			return nil
		},
	}

	// Add command-specific flags
	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground (don't detach)")

	return cmd
}
