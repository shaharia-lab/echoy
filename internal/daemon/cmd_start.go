package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewStartCmd creates a command to run the daemon
// Simplified parameters: Removed duplicate logger 'l' and unused 'webUIStaticDirectory'
func NewStartCmd(appConf config.Config, appConfig *config.AppConfig, log *logger.Logger, themeManager *theme.Manager, socketPath string) *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Echoy daemon",
		Long:  `Starts the Echoy daemon process that listens for commands via a Unix socket.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appConf.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(),
					appConfig,
					"daemon.start.attempt", // More specific event name
					telemetry.SeverityInfo, "Attempting to start daemon",
					nil,
				)
			}

			if !foreground {
				log.Info("Attempting to start daemon in background...")
				if isRunning, _ := isDaemonRunning(socketPath, log); isRunning {
					msg := "Daemon is already running"
					log.Info(msg)
					themeManager.GetCurrentTheme().Info().Println(msg)
					return nil
				}

				execPath, err := os.Executable()
				if err != nil {
					log.WithField("error", err).Error("failed to get executable path")
					return fmt.Errorf("failed to get executable path: %w", err)
				}

				daemonCmd := exec.Command(execPath, "start", "--foreground")
				daemonCmd.Stdout = nil
				daemonCmd.Stderr = nil
				daemonCmd.Stdin = nil

				setPlatformProcAttr(daemonCmd)

				if err := daemonCmd.Start(); err != nil {
					log.WithField("error", err).Error("Failed to start daemon process in background")
					themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon process: %v", err))
					return fmt.Errorf("failed to start daemon process: %w", err)
				}

				// Send telemetry after successful background start attempt
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(),
						appConfig,
						"daemon.start.background.success",
						telemetry.SeverityInfo, "Daemon started in background",
						nil,
					)
				}

				successMsg := fmt.Sprintf("Daemon starting in background mode (PID: %d). Listening on %s", daemonCmd.Process.Pid, socketPath)
				log.Info(successMsg)
				themeManager.GetCurrentTheme().Success().Println(successMsg)
				return nil
			}

			log.Info("Starting daemon in foreground mode...")

			// Create Daemon Configuration using the refactored package
			daemonCfg := Config{
				SocketPath:         socketPath,
				Logger:             nil,
				ShutdownTimeout:    30 * time.Second,
				ReadTimeout:        10 * time.Second,
				WriteTimeout:       10 * time.Second,
				CommandExecTimeout: 5 * time.Second,
			}
			daemonInstance := NewDaemon(daemonCfg)

			// ** IMPORTANT: Register command handlers here or elsewhere **
			// Example: daemonInstance.RegisterCommand("YOUR_COMMAND", yourCommandHandler)
			// You'll need to implement handlers for tasks like starting webservers, etc.

			// Setup graceful shutdown using context triggered by signals
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop() // Important: releases resources associated with NotifyContext

			// Start the daemon in a separate goroutine so we can wait for the context
			var startErr error
			daemonStopped := make(chan struct{})
			go func() {
				defer close(daemonStopped)
				if err := daemonInstance.Start(); err != nil {
					log.WithField("error", err).Error("Daemon failed to start")
					startErr = err
					stop()
				}
			}()

			// Check immediately if Start failed (e.g., socket in use)
			// Give Start a very brief moment to potentially fail fast
			select {
			case <-time.After(100 * time.Millisecond):
				log.WithField("socket", daemonCfg.SocketPath).Info("Daemon successfully started and listening")
				themeManager.GetCurrentTheme().Success().Printf("Daemon started and listening on %s\n", daemonCfg.SocketPath)
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(),
						appConfig,
						"daemon.start.foreground.success",
						telemetry.SeverityInfo, "Daemon started in foreground",
						nil,
					)
				}
			case <-ctx.Done(): // If stop() was called due to immediate Start error
				if startErr != nil {
					themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon: %v", startErr))
					return startErr
				}
				// Should not happen unless signal received instantly
				log.Info("Daemon startup interrupted by signal")
				return errors.New("daemon startup interrupted")
			}

			// Wait for shutdown signal
			<-ctx.Done()

			// Stop was called by the signal handler (or Start error), context is cancelled
			log.Info("Shutdown signal received, stopping daemon...")
			themeManager.GetCurrentTheme().Info().Println("Shutting down daemon...")

			// Initiate graceful shutdown
			daemonInstance.Stop() // This now handles timeouts internally

			// Wait for the daemon Start goroutine to fully exit (after Stop completes)
			<-daemonStopped
			log.Info("Daemon stopped gracefully.")
			themeManager.GetCurrentTheme().Success().Println("Daemon stopped.")

			// Return the start error if it occurred
			if startErr != nil {
				return fmt.Errorf("daemon exited with error: %w", startErr)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground (don't detach)")

	return cmd
}

// isDaemonRunning checks if the daemon is running by pinging its socket.
// Pass the logger for logging connection attempts/errors.
func isDaemonRunning(socketPath string, logger *logger.Logger) (bool, error) {
	logger.WithField("socket", socketPath).Debug("Checking if daemon is running")
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second) // Shorter timeout for check
	if err != nil {
		logger.WithField("error", err).Debug("Daemon not running or socket unavailable")
		return false, nil
	}
	defer conn.Close()

	if err = conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
		logger.WithField("error", err).Debug("Failed to set write deadline for ping")
		return false, err
	}
	_, err = conn.Write([]byte("PING\n"))
	if err != nil {
		logger.WithField("error", err).Debug("Failed to send PING to daemon")
		return false, err
	}

	// Read response
	if err = conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		logger.WithField("error", err).Debug("Failed to set read deadline for pong")
		return false, err
	}
	buffer := make([]byte, 128)
	n, err := conn.Read(buffer)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Warn("Daemon did not respond to PING within timeout")
		} else {
			logger.WithField("error", err).Warn("Failed to read PONG from daemon")
		}
		return false, err
	}

	response := string(buffer[:n])
	trimmedResponse := strings.TrimSpace(response)
	logger.WithField("response", trimmedResponse).Debug("Daemon response to PING")

	if trimmedResponse == "PONG" {
		logger.Debug("Daemon responded PONG successfully")
		return true, nil
	}

	logger.WithField("response", trimmedResponse).Debug("Daemon responded to PING unexpectedly")
	return false, fmt.Errorf("unexpected response from daemon: %s", trimmedResponse)
}
