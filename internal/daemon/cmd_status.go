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
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

// NewStatusCmd creates a command to check the daemon status
func NewStatusCmd(config config.Config, appConfig *config.AppConfig, logger *logger.Logger, themeManager *theme.Manager, socketPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check the status of the Echoy daemon",
		Long:  `Checks if the Echoy daemon and its components are currently running.`,
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

			daemonStatus, daemonResponsive := checkDaemonStatus(socketPath)

			if daemonResponsive {
				fmt.Fprintln(w, fmt.Sprintf("daemon\trunning\t-"))
			} else {
				fmt.Fprintln(w, fmt.Sprintf("daemon\t%s\t-", daemonStatus))
				w.Flush()

				themeManager.GetCurrentTheme().Warning().Println("\nDaemon is not running. Start it with 'echoy daemon start'")
				return nil
			}

			webServerPort := "10222"
			webServerStatus, webServerDetails := checkWebServerStatus(webServerPort)
			fmt.Fprintln(w, fmt.Sprintf("webserver\t%s\t%s", webServerStatus, webServerDetails))

			w.Flush()

			if webServerStatus == "running" {
				themeManager.GetCurrentTheme().Success().Println("\nAll components are running correctly")
			} else {
				themeManager.GetCurrentTheme().Warning().Println("\nSome components are not running correctly")
			}

			return nil
		},
	}
	return cmd
}

// checkDaemonStatus verifies if the daemon is running and responsive
func checkDaemonStatus(socketPath string) (string, bool) {
	conn, err := net.DialTimeout("unix", socketPath, 2*time.Second)
	if err != nil {
		return "not running", false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("PING\n"))
	if err != nil {
		return "not responsive", false
	}

	buffer := make([]byte, 128)
	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		return "socket error", false
	}

	n, err := conn.Read(buffer)
	if err != nil || n == 0 {
		return "not responding", false
	}

	response := string(buffer[:n])
	if strings.TrimSpace(response) == "PONG" {
		return "running", true
	}

	return "unexpected response", false
}

// checkWebServerStatus verifies if the webserver is running
func checkWebServerStatus(port string) (string, string) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/ping", port))
	if err != nil {
		return "not running", "-"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "running", fmt.Sprintf("port %s", port)
	}

	return "error", fmt.Sprintf("status %d", resp.StatusCode)
}
