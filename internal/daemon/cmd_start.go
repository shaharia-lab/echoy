package daemon

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/llm"
	"github.com/shaharia-lab/echoy/internal/tools"
	"github.com/shaharia-lab/echoy/internal/webserver"
	"github.com/shaharia-lab/echoy/internal/webui"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	"net"
	"net/http"
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
	mcpTools "github.com/shaharia-lab/mcp-tools"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewStartCmd creates a command to run the daemon
func NewStartCmd(config config.Config, appConfig *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager, socketPath string, webUIStaticDirectory string, l *logger.Logger) *cobra.Command {
	var foreground bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Echoy daemon",
		Long:  `Starts the Echoy daemon that processes background tasks and client requests.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if config.UsageTracking.Enabled {
				telemetryEvent.SendTelemetryEvent(
					context.Background(),
					appConfig,
					"daemon.start",
					telemetry.SeverityInfo, "Starting daemon",
					nil,
				)
			}

			logger.Info("Starting daemon...")
			defer logger.Sync()

			if !foreground {
				if isRunning, _ := isDaemonRunning(socketPath); isRunning {
					msg := "Daemon is already running"
					logger.Info(msg)
					themeManager.GetCurrentTheme().Info().Println(msg)
					return nil
				}

				execPath, err := os.Executable()
				if err != nil {
					return fmt.Errorf("failed to get executable path: %w", err)
				}

				daemonCmd := exec.Command(execPath, "start", "--foreground")

				daemonCmd.Stdout = nil
				daemonCmd.Stderr = nil
				daemonCmd.Stdin = nil

				setPlatformProcAttr(daemonCmd)

				if err := daemonCmd.Start(); err != nil {
					return fmt.Errorf("failed to start daemon process: %w", err)
				}

				themeManager.GetCurrentTheme().Success().Println("Daemon started in background mode")
				return nil
			}

			// From here on, we're in foreground mode
			daemon := NewDaemon(socketPath)

			ts := []mcp.Tool{
				mcpTools.GetWeather,
			}

			llmService, err := llm.NewLLMService(config.LLM)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create LLM service: %v", err))
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to create LLM service: %v", err))
				return err
			}

			historyService := goai.NewInMemoryChatHistoryStorage()

			chatService := chat.NewChatService(llmService, historyService)
			chatHandler := chat.NewChatHandler(chatService)
			webUIDownloaderHttpClient := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return nil
				},
			}

			ws := webserver.NewWebServer(
				"10222",
				webUIStaticDirectory,
				tools.NewProvider(ts),
				llm.NewLLMHandler(llm.GetSupportedLLMProviders()),
				chatHandler,
				webui.NewFrontendGitHubReleaseDownloader(webUIStaticDirectory, webUIDownloaderHttpClient, l),
			)
			daemon.WithWebServer(ws)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigChan
				logger.Info(fmt.Sprintf("Received signal %s, shutting down...", sig))
				cancel()
			}()

			if err := daemon.Start(); err != nil {
				logger.Error(fmt.Sprintf("Failed to start daemon: %v", err))
				themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to start daemon: %v", err))
				return err
			}

			themeManager.GetCurrentTheme().Success().Printf("Daemon started and listening on %s\n", daemon.SocketPath)
			logger.Info(fmt.Sprintf("Daemon started and listening on %s", daemon.SocketPath))

			// Wait for cancellation
			<-ctx.Done()
			logger.Info("Stopping daemon...")
			daemon.Stop()
			logger.Info("Daemon stopped")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&foreground, "foreground", "f", false, "Run daemon in foreground (don't detach)")

	return cmd
}

// isDaemonRunning checks if the daemon is currently running by attempting to connect to its socket
func isDaemonRunning(socketPath string) (bool, error) {
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return false, nil
	}
	defer conn.Close()

	// Try to ping the daemon
	_, err = conn.Write([]byte("PING\n"))
	if err != nil {
		return false, err
	}

	// Read response
	buffer := make([]byte, 128)
	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		return false, err
	}

	n, err := conn.Read(buffer)
	if err != nil {
		return false, err
	}

	response := string(buffer[:n])
	return strings.TrimSpace(response) == "PONG", nil
}
