package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog" // Use slog directly or via your wrapper
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shaharia-lab/echoy/internal/config"
	// Assuming logger wrapper is removed or provides slog:
	// "github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewStartCmd creates a command to run the daemon
func NewStartCmd(appConf config.Config, appConfig *config.AppConfig, themeManager *theme.Manager, socketPath string) *cobra.Command {
	var foreground bool
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Echoy daemon",
		Long:  `Starts the Echoy daemon process that listens for commands via a Unix socket.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Send telemetry for start attempt
			if appConf.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(), appConfig, "daemon.start.attempt",
					telemetry.SeverityInfo, "Attempting to start daemon", nil,
				)
			}

			if !foreground {
				// --- Background Mode ---
				logger.Info("Attempting to start daemon in background...")
				if isRunning, _ := isDaemonRunning(socketPath, logger); isRunning {
					msg := "Daemon is already running"
					logger.Info(msg)
					themeManager.GetCurrentTheme().Info().Println(msg)
					return nil // Success (already running)
				}

				execPath, err := os.Executable()
				if err != nil {
					logger.Error("Failed to get executable path", "error", err)
					return fmt.Errorf("failed to get executable path: %w", err)
				}

				// Prepare command to re-run in foreground
				daemonCmd := exec.Command(execPath, "start", "--foreground")
				daemonCmd.Stdout = nil
				daemonCmd.Stderr = nil
				daemonCmd.Stdin = nil
				setPlatformProcAttr(daemonCmd) // Apply detach attributes

				// Start the background process
				if err := daemonCmd.Start(); err != nil {
					logger.Error("Failed to start daemon process in background", "error", err)
					themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon process: %v", err))
					return fmt.Errorf("failed to start daemon process: %w", err)
				}

				// Send telemetry after successful background start initiation
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(), appConfig, "daemon.start.background.success",
						telemetry.SeverityInfo, "Daemon started in background", nil,
					)
				}

				// Success message for background launch
				pid := -1
				if daemonCmd.Process != nil {
					pid = daemonCmd.Process.Pid
				}
				successMsg := fmt.Sprintf("Daemon starting in background mode (PID: %d). Listening on %s", pid, socketPath)
				logger.Info(successMsg)
				themeManager.GetCurrentTheme().Success().Println(successMsg)
				return nil // Indicate successful launch
			}

			// --- Foreground Mode ---
			logger.Info("Starting daemon in foreground mode...")

			// Create Daemon Configuration
			daemonCfg := Config{
				SocketPath:         socketPath,
				Logger:             logger,           // Pass the actual slog logger
				ShutdownTimeout:    30 * time.Second, // Example defaults, consider overriding from appConf
				ReadTimeout:        10 * time.Second,
				WriteTimeout:       10 * time.Second,
				CommandExecTimeout: 5 * time.Second,
				MaxConnections:     100, // Example connection limit
			}
			// Optional: Override daemonCfg fields from appConf here

			daemonInstance := NewDaemon(daemonCfg)

			// *** IMPORTANT: Register application-specific command handlers here ***
			// Example: daemonInstance.RegisterCommand("START_WEBSERVER", handleStartWebServer)
			//          daemonInstance.RegisterCommand("STOP_WEBSERVER", handleStopWebServer)
			//          ...etc...

			// Setup graceful shutdown context
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop() // Release context resources

			// --- Error Channel for Start Goroutine (Feedback #9) ---
			errChan := make(chan error, 1)
			daemonStopped := make(chan struct{}) // Signals when the Start goroutine finishes

			go func() {
				defer close(daemonStopped) // Signal exit when this goroutine returns
				err := daemonInstance.Start()
				if err != nil {
					logger.Error("Daemon failed to start", "error", err)
					// Send error non-blockingly or it might deadlock if main thread isn't reading
					select {
					case errChan <- err:
					default: // Avoid blocking if channel buffer is full (shouldn't be)
					}
					stop() // Trigger context cancellation on start failure
				}
				// If Start returns nil, it means setup was successful and listener loop started.
				// The loop will run until Stop() is called.
			}()

			// Wait briefly for initial start result or signal
			select {
			case err := <-errChan: // Check if Start failed immediately
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon: %v", err))
				return err // Return the specific startup error
			case <-time.After(200 * time.Millisecond): // Give Start a bit more time
				// Assume startup is okay if no error received quickly
				logger.Info("Daemon successfully started and listening", "socket", daemonCfg.SocketPath)
				themeManager.GetCurrentTheme().Success().Printf("Daemon started and listening on %s\n", daemonCfg.SocketPath)
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(), appConfig, "daemon.start.foreground.success",
						telemetry.SeverityInfo, "Daemon started in foreground", nil,
					)
				}
			case <-ctx.Done(): // Signal received *during* initial startup check
				// This case is less likely unless signal is sent immediately
				logger.Info("Daemon startup interrupted by signal before confirmation")
				// Wait for the Start goroutine to potentially finish/report error
				select {
				case err := <-errChan:
					return fmt.Errorf("daemon startup interrupted: %w", err)
				case <-daemonStopped: // Start goroutine exited cleanly or after error
					return errors.New("daemon startup interrupted by signal")
				}
			}

			// --- Wait for Shutdown Signal ---
			<-ctx.Done() // Block here until SIGINT/SIGTERM is received

			// Context is cancelled (by signal or potentially by Start error handler)
			logger.Info("Shutdown signal received or start failed, stopping daemon...")
			themeManager.GetCurrentTheme().Info().Println("Shutting down daemon...")

			// Initiate graceful shutdown (Stop has internal timeout)
			daemonInstance.Stop()

			// Wait for the daemon's Start goroutine to fully exit
			<-daemonStopped
			logger.Info("Daemon stopped gracefully.")
			themeManager.GetCurrentTheme().Success().Println("Daemon stopped.")

			// Check if an error occurred during Start after the initial check
			// (e.g., if stop() was called due to error after the 200ms timeout)
			select {
			case err := <-errChan:
				return fmt.Errorf("daemon exited with error: %w", err)
			default:
				// No error reported, clean shutdown
				return nil
			}
		},
	}

	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground (don't detach)")

	return cmd
}

// isDaemonRunning checks if the daemon is running by pinging its socket.
func isDaemonRunning(socketPath string, logger *slog.Logger) (bool, error) {
	logger.Debug("Checking if daemon is running", "socket", socketPath)
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		// Log level Debug is appropriate, as this is normal if daemon isn't running
		logger.Debug("Daemon check connection failed (likely not running)", "socket", socketPath, "error", err)
		return false, nil // Treat connection error as "not running"
	}
	defer conn.Close()

	// Ping the daemon
	if err = conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
		logger.Warn("Daemon check failed to set write deadline for ping", "socket", socketPath, "error", err)
		return false, err // Return error, might indicate socket issue
	}
	if _, err = conn.Write([]byte("PING\n")); err != nil {
		logger.Warn("Daemon check failed to send PING", "socket", socketPath, "error", err)
		return false, err // Error writing often means daemon issue
	}

	// Read response
	if err = conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		logger.Warn("Daemon check failed to set read deadline for pong", "socket", socketPath, "error", err)
		return false, err
	}
	buffer := make([]byte, 32) // Small buffer just for PONG
	n, err := conn.Read(buffer)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Warn("Daemon check timeout waiting for PONG", "socket", socketPath, "timeout", "1s")
		} else {
			logger.Warn("Daemon check failed to read PONG", "socket", socketPath, "error", err)
		}
		return false, err // No response or error reading
	}

	// Check response
	response := string(buffer[:n])
	trimmedResponse := strings.TrimSpace(response)
	logger.Debug("Daemon check received response", "socket", socketPath, "response", trimmedResponse)

	if trimmedResponse == "PONG" {
		logger.Debug("Daemon check successful (PONG received)", "socket", socketPath)
		return true, nil
	}

	// Unexpected response
	logger.Warn("Daemon check received unexpected response", "socket", socketPath, "response", trimmedResponse)
	return false, fmt.Errorf("unexpected response from daemon: %q", trimmedResponse)
}
