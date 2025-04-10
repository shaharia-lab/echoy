package daemon

import (
	"context"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

// IsDaemonRunning checks if the daemon is running by attempting to connect to its socket
func IsDaemonRunning(socketPath string, logger *slog.Logger) (bool, error) {
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		if logger != nil {
			logger.Debug("Daemon socket file does not exist", "path", socketPath)
		}
		return false, nil
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		if _, statErr := os.Stat(socketPath); statErr == nil {
			if rmErr := os.Remove(socketPath); rmErr != nil && logger != nil {
				logger.Warn("Failed to remove stale socket file", "path", socketPath, "error", rmErr)
			} else if logger != nil {
				logger.Debug("Removed stale socket file", "path", socketPath)
			}
		}
		return false, nil
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	provider := &UnixSocketProvider{
		SocketPath: socketPath,
		Timeout:    500 * time.Millisecond,
	}
	client := NewClient(provider, 1*time.Second, 2*time.Second)

	response, err := client.Execute(ctx, "PING", nil)
	if err != nil {
		if logger != nil {
			logger.Warn("Daemon not responding to PING", "error", err)
		}
		return false, nil
	}

	// Check if response is PONG or has OK: PONG prefix
	response = strings.TrimSpace(response)
	if response != "PONG" && response != "OK: PONG" {
		if logger != nil {
			logger.Warn("Unexpected response from daemon", "response", response)
		}
		return false, nil
	}

	if logger != nil {
		logger.Debug("Daemon is running", "path", socketPath)
	}
	return true, nil
}
