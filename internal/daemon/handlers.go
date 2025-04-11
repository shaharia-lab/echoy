package daemon

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/types"
	"sort"
	"strings"
)

// DefaultPingHandler is a simple ping handler that responds with "PONG".
func DefaultPingHandler(ctx context.Context, args []string) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("ping cancelled: %w", ctx.Err())
	default:
		return "PONG", nil
	}
}

// MakeDefaultStatusHandler creates a status handler closure capturing the daemon instance.
func MakeDefaultStatusHandler(d *Daemon) types.CommandFunc {
	return func(ctx context.Context, args []string) (string, error) {
		d.connMu.RLock()
		connCount := len(d.connections)
		d.connMu.RUnlock()

		d.cmdMu.RLock()
		cmdCount := len(d.commands)
		cmdNames := make([]string, 0, cmdCount)
		for name := range d.commands {
			cmdNames = append(cmdNames, name)
		}
		d.cmdMu.RUnlock()

		sort.Strings(cmdNames)

		select {
		case <-ctx.Done():
			return "", fmt.Errorf("status cancelled: %w", ctx.Err())
		default:
			status := fmt.Sprintf(
				"Connections: %d active (Limit: %d)\nCommands: %d registered (%s)",
				connCount,
				d.config.MaxConnections,
				cmdCount,
				strings.Join(cmdNames, ", "),
			)
			return status, nil
		}
	}
}

// MakeDefaultStopHandler creates a stop handler closure capturing the daemon instance.
func MakeDefaultStopHandler(d *Daemon) types.CommandFunc {
	return func(ctx context.Context, args []string) (string, error) {
		d.logger.Info("STOP command received via connection, triggering daemon shutdown.")
		go d.Stop()
		return "Daemon stop initiated.", nil
	}
}
