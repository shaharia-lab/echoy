package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/shaharia-lab/echoy/internal/config"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewStopCmd creates a command to stop the running daemon
func NewStopCmd(appConf config.Config, appConfig *config.AppConfig, logger *slog.Logger, themeManager *theme.Manager, socketPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the running Echoy daemon",
		Long:  `Sends a stop command to the running Echoy daemon to shut it down gracefully.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if appConf.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(), appConfig, "daemon.stop.attempt",
					telemetry.SeverityInfo, "Attempting to stop daemon", nil,
				)
			}

			logger.Info("Attempting to stop daemon...", "socket", socketPath)

			conn, err := net.DialTimeout("unix", socketPath, 3*time.Second) // Slightly shorter timeout for connect
			if err != nil {
				if errors.Is(err, os.ErrNotExist) || strings.Contains(err.Error(), "no such file or directory") {
					logger.Info("Daemon socket not found, daemon likely not running.", "socket", socketPath)
					themeManager.GetCurrentTheme().Info().Println("Daemon is not running.")
					return nil
				}

				errMsg := fmt.Sprintf("Failed to connect to daemon at %s", socketPath)
				logger.Error(errMsg, "error", err)
				themeManager.GetCurrentTheme().Error().Println(errMsg + fmt.Sprintf(": %v", err))
				return fmt.Errorf("connection failed: %w", err)
			}
			defer conn.Close()
			logger.Debug("Connected to daemon socket", "socket", socketPath)

			if err = conn.SetWriteDeadline(time.Now().Add(3 * time.Second)); err != nil {
				logger.Error("Failed to set write deadline for stop command", "error", err)
				themeManager.GetCurrentTheme().Error().Println("Failed to set write deadline.")
				return fmt.Errorf("set write deadline failed: %w", err)
			}
			_, err = conn.Write([]byte("STOP\n"))
			if err != nil {
				errMsg := "Failed to send STOP command to daemon"
				logger.Error(errMsg, "error", err)
				themeManager.GetCurrentTheme().Error().Println(errMsg + fmt.Sprintf(": %v", err))
				return fmt.Errorf("failed to send command: %w", err)
			}
			logger.Debug("STOP command sent to daemon")

			readTimeout := 5 * time.Second
			if err = conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				logger.Error("Failed to set read deadline for response", "error", err)
				themeManager.GetCurrentTheme().Error().Println("Failed to set read deadline for response.")
				return fmt.Errorf("set read deadline failed: %w", err)
			}

			buffer := make([]byte, 256)
			n, readErr := conn.Read(buffer)

			finalMessage := "Stop command sent successfully. Daemon shutdown initiated."
			isSuccess := true

			if readErr != nil {
				if errors.Is(readErr, io.EOF) || errors.Is(readErr, net.ErrClosed) || strings.Contains(readErr.Error(), "use of closed network connection") {
					logger.Debug("Daemon closed connection after STOP command (expected).")
				} else if errors.Is(readErr, os.ErrDeadlineExceeded) {
					logger.Warn("Timeout waiting for daemon response/connection close after STOP.", "timeout", readTimeout)
					finalMessage = "Stop command sent, but no confirmation received within timeout."
				} else {
					errMsg := "Error reading response from daemon after STOP"
					logger.Error(errMsg, "error", readErr)
					themeManager.GetCurrentTheme().Error().Println(errMsg + fmt.Sprintf(": %v", readErr))
					isSuccess = false
					finalMessage = "Error receiving confirmation from daemon."
					return fmt.Errorf("failed reading daemon response: %w", readErr)
				}
			} else {
				response := string(buffer[:n])
				trimmedResponse := strings.TrimSpace(response)
				logger.Debug("Received response from daemon", "response", trimmedResponse)
				if strings.HasPrefix(trimmedResponse, "OK:") {
					logger.Debug("Daemon acknowledged STOP command.")
				} else if strings.HasPrefix(trimmedResponse, "ERROR: unknown command 'STOP'") {
					errMsg := "Daemon reported 'STOP' is an unknown command (handler not registered?)"
					logger.Error(errMsg)
					themeManager.GetCurrentTheme().Error().Println(errMsg)
					isSuccess = false
					finalMessage = errMsg
					return errors.New(errMsg)
				} else {
					errMsg := "Received unexpected response from daemon after STOP"
					logger.Warn(errMsg, "response", trimmedResponse)
					finalMessage = "Stop command sent, but received unexpected response."
				}
			}

			if isSuccess {
				themeManager.GetCurrentTheme().Success().Println(finalMessage)
				logger.Info(finalMessage)
				if appConf.UsageTracking.Enabled {
					telemetryEvent.SendTelemetryEvent(
						context.Background(), appConfig, "daemon.stop.success",
						telemetry.SeverityInfo, finalMessage, nil,
					)
				}
				return nil
			} else {
				themeManager.GetCurrentTheme().Error().Println(finalMessage)
				logger.Error(finalMessage)
				return errors.New(finalMessage)
			}
		},
	}

	return cmd
}
