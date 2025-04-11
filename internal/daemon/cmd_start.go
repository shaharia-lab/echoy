package daemon

import (
	"context"
	"errors"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/webserver"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"

	"github.com/shaharia-lab/echoy/internal/config"
	loggerInt "github.com/shaharia-lab/echoy/internal/logger"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
)

// NewStartCmd creates a command to run the daemon
func NewStartCmd(container *cli.Container, appConf config.Config, appConfig *config.AppConfig, themeManager *theme.Manager, socketPath string, webUIStaticDirectory string, sLogger *slog.Logger) *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Echoy daemon",
		Long:  `Starts the Echoy daemon process that listens for commands via a Unix socket.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer container.Logger.Flush()

			if container.ConfigFromFile.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(), appConfig, "daemon.start.attempt",
					telemetry.SeverityInfo, "Attempting to start daemon", nil,
				)
			}

			daemonLog, err := loggerInt.NewZapLogger(loggerInt.Config{
				LogLevel:    loggerInt.DebugLevel,
				LogFilePath: fmt.Sprintf("%s/daemon.log", container.Paths[filesystem.LogsDirectory]),
				MaxSizeMB:   50,
				MaxAgeDays:  14,
				MaxBackups:  5,
			})
			if err != nil {
				panic(fmt.Sprintf("Failed to initialize logger: %v", err))
			}

			if !foreground {
				container.Logger.WithFields(map[string]interface{}{
					"socket":  socketPath,
					"pid":     os.Getpid(),
					"command": "start",
				}).Info("Attempting to start daemon in background...")

				if isRunning, err := isDaemonRunning(socketPath, container.Logger); isRunning {
					if err != nil {
						container.Logger.WithFields(map[string]interface{}{
							loggerInt.ErrorKey: err,
							"command":          "start",
							"socket":           socketPath,
						}).Error("Failed to check if daemon is running")

						return fmt.Errorf("failed to check if daemon is running: %w", err)
					}

					msg := "Daemon is already running"

					container.Logger.WithFields(map[string]interface{}{
						"socket":  socketPath,
						"command": "start",
					}).Info("Daemon is already running")

					themeManager.GetCurrentTheme().Info().Println(msg)
					return nil
				}

				execPath, err := os.Executable()
				if err != nil {
					container.Logger.WithFields(map[string]interface{}{
						loggerInt.ErrorKey: err,
						"command":          "start",
						"socket":           socketPath,
					}).Error("Failed to get executable path")

					return fmt.Errorf("failed to get executable path: %w", err)
				}

				daemonCmd := exec.Command(execPath, "start", "--foreground")
				daemonCmd.Stdout = nil
				daemonCmd.Stderr = nil
				daemonCmd.Stdin = nil
				setPlatformProcAttr(daemonCmd)

				if err := daemonCmd.Start(); err != nil {
					container.Logger.WithFields(map[string]interface{}{
						loggerInt.ErrorKey: err,
						"command":          "start",
						"socket":           socketPath,
					}).Error("Failed to start daemon process in background")

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
				container.Logger.WithFields(map[string]interface{}{
					"socket":     socketPath,
					"daemon_pid": pid,
					"command":    "start",
				}).Info("Daemon starting in background mode")

				themeManager.GetCurrentTheme().Success().Println(successMsg)
				return nil
			}

			container.Logger.WithFields(map[string]interface{}{
				"socket":  socketPath,
				"command": "start",
			}).Info("Starting daemon in foreground mode...")

			webSrvr, err := webserver.BuildWebserver(appConf, themeManager, webUIStaticDirectory, container.Paths[filesystem.LogsDirectory])
			if err != nil {
				container.Logger.WithFields(map[string]interface{}{
					loggerInt.ErrorKey: err,
					"command":          "start",
					"socket":           socketPath,
				}).Error("Failed to build web server")

				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to build web server: %v", err))
				return fmt.Errorf("failed to build web server: %w", err)
			}

			daemonCfg := Config{
				SocketPath:         socketPath,
				Logger:             daemonLog,
				ShutdownTimeout:    30 * time.Second,
				ReadTimeout:        10 * time.Second,
				WriteTimeout:       10 * time.Second,
				CommandExecTimeout: 5 * time.Second,
				MaxConnections:     100,
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			daemonInstance := NewDaemon(daemonCfg, daemonLog)
			daemonInstance.SetCancelFunc(stop)

			daemonInstance.RegisterCommand("PING", DefaultPingHandler)
			daemonInstance.RegisterCommand("STATUS", MakeDefaultStatusHandler(daemonInstance))
			daemonInstance.RegisterCommand("STOP", MakeDefaultStopHandler(daemonInstance))
			daemonInstance.RegisterCommand("WEBSERVER", webSrvr.DaemonCommandHandler())

			errChan := make(chan error, 1)
			daemonStopped := make(chan struct{})

			go func() {
				defer close(daemonStopped)
				err := daemonInstance.Start()
				if err != nil {
					container.Logger.WithFields(map[string]interface{}{
						loggerInt.ErrorKey: err,
						"command":          "start",
						"socket":           socketPath,
					}).Error("Daemon failed to start")

					select {
					case errChan <- err:
					default:
					}
					stop()
				}
			}()

			select {
			case err := <-errChan:
				container.Logger.WithFields(map[string]interface{}{
					loggerInt.ErrorKey: err,
					"command":          "start",
					"socket":           socketPath,
				}).Error("Daemon failed to start")

				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon: %v", err))
				return err
			case <-time.After(200 * time.Millisecond):
				container.Logger.WithFields(map[string]interface{}{
					"socket":  socketPath,
					"command": "start",
				}).Info("Daemon started successfully and listening...")

				themeManager.GetCurrentTheme().Success().Printf("Daemon started and listening on %s\n", daemonCfg.SocketPath)
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(), appConfig, "daemon.start.foreground.success",
						telemetry.SeverityInfo, "Daemon started in foreground", nil,
					)
				}
			case <-ctx.Done():
				container.Logger.WithFields(map[string]interface{}{
					"socket":  socketPath,
					"command": "start",
				}).Info("Daemon startup interrupted by signal")

				select {
				case err := <-errChan:
					container.Logger.WithFields(map[string]interface{}{
						loggerInt.ErrorKey: err,
						"command":          "start",
						"socket":           socketPath,
						"daemon":           "received_stopped",
					}).Error("Daemon startup interrupted")

					return fmt.Errorf("daemon startup interrupted: %w", err)
				case <-daemonStopped:
					container.Logger.WithFields(map[string]interface{}{
						"socket":  socketPath,
						"command": "start",
						"daemon":  "stopped",
					}).Info("Daemon startup interrupted by signal")

					return errors.New("daemon startup interrupted by signal")
				}
			}

			<-ctx.Done()

			container.Logger.WithFields(map[string]interface{}{
				"socket":  socketPath,
				"command": "start",
				"daemon":  "stopped",
			}).Info("Shutdown signal received or start failed, stopping daemon...")

			themeManager.GetCurrentTheme().Info().Println("Shutting down daemon...")

			daemonInstance.Stop()

			<-daemonStopped

			container.Logger.WithFields(map[string]interface{}{
				"socket":  socketPath,
				"command": "start",
				"daemon":  "stopped",
			}).Info("Daemon stopped gracefully.")

			themeManager.GetCurrentTheme().Success().Println("Daemon stopped.")

			select {
			case err := <-errChan:
				container.Logger.WithFields(map[string]interface{}{
					loggerInt.ErrorKey: err,
					"command":          "start",
					"socket":           socketPath,
					"daemon":           "stopped",
				}).Error("Daemon stopped with error")

				return fmt.Errorf("daemon exited with error: %w", err)
			default:
				return nil
			}
		},
	}

	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground (don't detach)")

	return cmd
}

func isDaemonRunning(socketPath string, logger loggerInt.Logger) (bool, error) {
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
