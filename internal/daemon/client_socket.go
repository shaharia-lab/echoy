package daemon

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"time"
)

// Commander defines an interface for sending commands to the daemon
type Commander interface {
	// Execute sends a command to the daemon and returns the response
	Execute(ctx context.Context, cmd string) (string, error)
	// IsRunning checks if the daemon is running and responsive
	IsRunning(ctx context.Context) (bool, string)
}

// ConnectionProvider defines an interface for creating connections to the daemon
type ConnectionProvider interface {
	// Connect establishes a connection to the daemon
	Connect(ctx context.Context) (net.Conn, error)
}

// UnixSocketProvider provides connections to a Unix socket
type UnixSocketProvider struct {
	SocketPath string
	Timeout    time.Duration
}

// Connect implements ConnectionProvider.Connect
func (p *UnixSocketProvider) Connect(ctx context.Context) (net.Conn, error) {
	// We use DialTimeout instead of ctx to maintain compatibility with net.Dial
	conn, err := net.DialTimeout("unix", p.SocketPath, p.Timeout)
	if err != nil {
		return nil, err
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		if ctx.Err() != nil {
			conn.Close()
		}
	}()

	return conn, nil
}

// DaemonClient implements Commander using a ConnectionProvider
type DaemonClient struct {
	Provider     ConnectionProvider
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewDaemonClient creates a new DaemonClient with a UnixSocketProvider
func NewDaemonClient(socketPath string) *DaemonClient {
	return &DaemonClient{
		Provider: &UnixSocketProvider{
			SocketPath: socketPath,
			Timeout:    2 * time.Second,
		},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

// Execute implements Commander.Execute
func (c *DaemonClient) Execute(ctx context.Context, cmd string) (string, error) {
	conn, err := c.Provider.Connect(ctx)
	if err != nil {
		return "", errors.New("failed to connect to daemon: " + err.Error())
	}
	defer conn.Close()

	// Set write deadline
	if c.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout)); err != nil {
			return "", errors.New("failed to set write deadline: " + err.Error())
		}
	}

	// Ensure command ends with newline
	if !strings.HasSuffix(cmd, "\n") {
		cmd = cmd + "\n"
	}

	// Send command
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", errors.New("failed to send command: " + err.Error())
	}

	// Set read deadline
	if c.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
			return "", errors.New("failed to set read deadline: " + err.Error())
		}
	}

	// Read response
	reader := bufio.NewReader(conn)
	var response strings.Builder

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF || response.Len() > 0 {
					// Return what we have if we've read anything
					return strings.TrimSpace(response.String()), nil
				}
				return "", errors.New("failed to read response: " + err.Error())
			}

			response.WriteString(line)

			// Check for end marker
			trimmed := strings.TrimSpace(line)
			if trimmed == "END" || trimmed == "" {
				return strings.TrimSpace(response.String()), nil
			}
		}
	}
}

// IsRunning implements Commander.IsRunning
func (c *DaemonClient) IsRunning(ctx context.Context) (bool, string) {
	response, err := c.Execute(ctx, "PING")
	if err != nil {
		return false, "not running: " + err.Error()
	}

	if strings.TrimSpace(response) == "PONG" {
		return true, "running"
	}

	return false, "unexpected response: " + response
}
