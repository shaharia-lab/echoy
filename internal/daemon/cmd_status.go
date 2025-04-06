package daemon

import (
	"context"
	"fmt"
	"github.com/olekukonko/tablewriter"
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

			// Initialize table
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Component", "Status", "Details"})
			table.SetBorder(false)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
			table.SetHeaderColor(
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
				tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
			)

			// Check daemon status
			daemonStatus, daemonResponsive := checkDaemonStatus(socketPath)

			// Set daemon status color based on state
			var daemonStatusColor []tablewriter.Colors
			if daemonResponsive {
				daemonStatusColor = []tablewriter.Colors{
					{},
					{tablewriter.Bold, tablewriter.FgGreenColor},
					{},
				}
				table.Rich([]string{"Daemon", "Running", "-"}, daemonStatusColor)
			} else {
				daemonStatusColor = []tablewriter.Colors{
					{},
					{tablewriter.Bold, tablewriter.FgRedColor},
					{},
				}
				table.Rich([]string{"Daemon", daemonStatus, "-"}, daemonStatusColor)
				table.Render()
				return nil
			}

			// Check webserver status only if daemon is running
			webServerPort := "10222" // Default port from cmd_start.go
			webServerStatus, webServerDetails := checkWebServerStatus(webServerPort)

			// Set webserver status color based on state
			var webServerStatusColor []tablewriter.Colors
			if webServerStatus == "Running" {
				webServerStatusColor = []tablewriter.Colors{
					{},
					{tablewriter.Bold, tablewriter.FgGreenColor},
					{},
				}
			} else {
				webServerStatusColor = []tablewriter.Colors{
					{},
					{tablewriter.Bold, tablewriter.FgRedColor},
					{},
				}
			}

			table.Rich([]string{"WebServer", webServerStatus, webServerDetails}, webServerStatusColor)

			// Render the table
			table.Render()

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
