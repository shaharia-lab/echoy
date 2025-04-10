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
	conn, err := net.DialTimeout("unix", p.SocketPath, p.Timeout)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		if ctx.Err() != nil {
			conn.Close()
		}
	}()

	return conn, nil
}

// Client implements Commander using a ConnectionProvider
type Client struct {
	Provider     ConnectionProvider
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewClient creates a new DaemonClient with a UnixSocketProvider
func NewClient(connectionProvider ConnectionProvider, readTimeout, writeTimeout time.Duration) *Client {
	return &Client{
		Provider:     connectionProvider,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}

// Execute implements Commander.Execute
func (c *Client) Execute(ctx context.Context, cmd string) (string, error) {
	conn, err := c.Provider.Connect(ctx)
	if err != nil {
		return "", errors.New("failed to connect to daemon: " + err.Error())
	}
	defer conn.Close()

	if c.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout)); err != nil {
			return "", errors.New("failed to set write deadline: " + err.Error())
		}
	}

	if !strings.HasSuffix(cmd, "\n") {
		cmd = cmd + "\n"
	}

	if _, err := conn.Write([]byte(cmd)); err != nil {
		return "", errors.New("failed to send command: " + err.Error())
	}

	if c.ReadTimeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(c.ReadTimeout)); err != nil {
			return "", errors.New("failed to set read deadline: " + err.Error())
		}
	}

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
					return strings.TrimSpace(response.String()), nil
				}
				return "", errors.New("failed to read response: " + err.Error())
			}

			response.WriteString(line)

			trimmed := strings.TrimSpace(line)
			if trimmed == "END" || trimmed == "" {
				return strings.TrimSpace(response.String()), nil
			}
		}
	}
}

// IsRunning implements Commander.IsRunning
func (c *Client) IsRunning(ctx context.Context) (bool, string) {
	response, err := c.Execute(ctx, "PING")
	if err != nil {
		return false, "not running: " + err.Error()
	}

	if strings.TrimSpace(response) == "PONG" {
		return true, "running"
	}

	return false, "unexpected response: " + response
}
