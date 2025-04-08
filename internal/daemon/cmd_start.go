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
			if appConf.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(), appConfig, "daemon.start.attempt",
					telemetry.SeverityInfo, "Attempting to start daemon", nil,
				)
			}

			if !foreground {
				logger.Info("Attempting to start daemon in background...")
				if isRunning, _ := isDaemonRunning(socketPath, logger); isRunning {
					msg := "Daemon is already running"
					logger.Info(msg)
					themeManager.GetCurrentTheme().Info().Println(msg)
					return nil
				}

				execPath, err := os.Executable()
				if err != nil {
					logger.Error("Failed to get executable path", "error", err)
					return fmt.Errorf("failed to get executable path: %w", err)
				}

				daemonCmd := exec.Command(execPath, "start", "--foreground")
				daemonCmd.Stdout = nil
				daemonCmd.Stderr = nil
				daemonCmd.Stdin = nil
				setPlatformProcAttr(daemonCmd)

				if err := daemonCmd.Start(); err != nil {
					logger.Error("Failed to start daemon process in background", "error", err)
					themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon process: %v", err))
					return fmt.Errorf("failed to start daemon process: %w", err)
				}

				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(), appConfig, "daemon.start.background.success",
						telemetry.SeverityInfo, "Daemon started in background", nil,
					)
				}

				pid := -1
				if daemonCmd.Process != nil {
					pid = daemonCmd.Process.Pid
				}
				successMsg := fmt.Sprintf("Daemon starting in background mode (PID: %d). Listening on %s", pid, socketPath)
				logger.Info(successMsg)
				themeManager.GetCurrentTheme().Success().Println(successMsg)
				return nil
			}

			logger.Info("Starting daemon in foreground mode...")

			daemonCfg := Config{
				SocketPath:         socketPath,
				Logger:             logger,
				ShutdownTimeout:    30 * time.Second,
				ReadTimeout:        10 * time.Second,
				WriteTimeout:       10 * time.Second,
				CommandExecTimeout: 5 * time.Second,
				MaxConnections:     100,
			}

			daemonInstance := NewDaemon(daemonCfg)

			daemonInstance.RegisterCommand("PING", DefaultPingHandler)
			daemonInstance.RegisterCommand("STATUS", MakeDefaultStatusHandler(daemonInstance))
			daemonInstance.RegisterCommand("STOP", MakeDefaultStopHandler(daemonInstance))

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			errChan := make(chan error, 1)
			daemonStopped := make(chan struct{})

			go func() {
				defer close(daemonStopped)
				err := daemonInstance.Start()
				if err != nil {
					logger.Error("Daemon failed to start", "error", err)
					select {
					case errChan <- err:
					default:
					}
					stop()
				}
			}()

			select {
			case err := <-errChan:
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon: %v", err))
				return err
			case <-time.After(200 * time.Millisecond):
				logger.Info("Daemon successfully started and listening", "socket", daemonCfg.SocketPath)
				themeManager.GetCurrentTheme().Success().Printf("Daemon started and listening on %s\n", daemonCfg.SocketPath)
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(), appConfig, "daemon.start.foreground.success",
						telemetry.SeverityInfo, "Daemon started in foreground", nil,
					)
				}
			case <-ctx.Done():
				logger.Info("Daemon startup interrupted by signal before confirmation")
				select {
				case err := <-errChan:
					return fmt.Errorf("daemon startup interrupted: %w", err)
				case <-daemonStopped:
					return errors.New("daemon startup interrupted by signal")
				}
			}

			<-ctx.Done()

			logger.Info("Shutdown signal received or start failed, stopping daemon...")
			themeManager.GetCurrentTheme().Info().Println("Shutting down daemon...")

			daemonInstance.Stop()

			<-daemonStopped
			logger.Info("Daemon stopped gracefully.")
			themeManager.GetCurrentTheme().Success().Println("Daemon stopped.")

			select {
			case err := <-errChan:
				return fmt.Errorf("daemon exited with error: %w", err)
			default:
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
		logger.Debug("Daemon check connection failed (likely not running)", "socket", socketPath, "error", err)
		return false, nil
	}
	defer conn.Close()

	if err = conn.SetWriteDeadline(time.Now().Add(1 * time.Second)); err != nil {
		logger.Warn("Daemon check failed to set write deadline for ping", "socket", socketPath, "error", err)
		return false, err
	}
	if _, err = conn.Write([]byte("PING\n")); err != nil {
		logger.Warn("Daemon check failed to send PING", "socket", socketPath, "error", err)
		return false, err
	}

	if err = conn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
		logger.Warn("Daemon check failed to set read deadline for pong", "socket", socketPath, "error", err)
		return false, err
	}
	buffer := make([]byte, 32)
	n, err := conn.Read(buffer)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			logger.Warn("Daemon check timeout waiting for PONG", "socket", socketPath, "timeout", "1s")
		} else {
			logger.Warn("Daemon check failed to read PONG", "socket", socketPath, "error", err)
		}
		return false, err
	}

	response := string(buffer[:n])
	trimmedResponse := strings.TrimSpace(response)
	logger.Debug("Daemon check received response", "socket", socketPath, "response", trimmedResponse)

	if trimmedResponse == "PONG" {
		logger.Debug("Daemon check successful (PONG received)", "socket", socketPath)
		return true, nil
	}

	logger.Warn("Daemon check received unexpected response", "socket", socketPath, "response", trimmedResponse)
	return false, fmt.Errorf("unexpected response from daemon: %q", trimmedResponse)
}
