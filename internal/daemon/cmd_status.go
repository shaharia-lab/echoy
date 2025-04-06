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

			// Prepare table writer
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "COMPONENT\tSTATUS\tDETAILS")
			fmt.Fprintln(w, "---------\t------\t-------")

			// Check daemon status
			daemonStatus, daemonResponsive := checkDaemonStatus(socketPath)

			// Set daemon status based on responsiveness
			if daemonResponsive {
				fmt.Fprintln(w, "Daemon\tRunning\t-")
			} else {
				fmt.Fprintln(w, fmt.Sprintf("Daemon\t%s\t-", daemonStatus))
				w.Flush()
				return nil
			}

			// Check webserver status only if daemon is running
			webServerPort := "10222" // Default port from cmd_start.go
			webServerStatus, webServerDetails := checkWebServerStatus(webServerPort)
			fmt.Fprintln(w, fmt.Sprintf("WebServer\t%s\t%s", webServerStatus, webServerDetails))

			// Flush the tabwriter
			w.Flush()

			// Log complete status
			if daemonResponsive && webServerStatus == "Running" {
				logger.Info("All components are running properly")
			} else {
				logger.Warn("Some components are not running properly")
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
		return "Not running", false
	}
	defer conn.Close()

	_, err = conn.Write([]byte("PING\n"))
	if err != nil {
		return "Socket exists but not responsive", false
	}

	// Read response
	buffer := make([]byte, 128)
	err = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err != nil {
		return "Socket error", false
	}

	n, err := conn.Read(buffer)
	if err != nil || n == 0 {
		return "Not responding to commands", false
	}

	response := string(buffer[:n])
	if strings.TrimSpace(response) == "PONG" {
		return "Running", true
	}

	return "Unexpected response", false
}

// checkWebServerStatus verifies if the webserver is running
func checkWebServerStatus(port string) (string, string) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/ping", port))
	if err != nil {
		return "Not running", "-"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "Running", fmt.Sprintf("Port %s", port)
	}

	return "Error", fmt.Sprintf("Returned status %d", resp.StatusCode)
}
