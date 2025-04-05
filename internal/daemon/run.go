package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// DefaultSocketPath is the default location for the daemon's Unix socket
const DefaultSocketPath = "$HOME/.echoy/echoy.sock"

// Daemon represents a Unix socket daemon service
type Daemon struct {
	SocketPath  string
	listener    net.Listener
	stopChan    chan struct{}
	connections map[net.Conn]struct{}
}

// NewDaemon creates a new Daemon instance with default configuration
func NewDaemon() *Daemon {
	socketPath := ResolveSocketPath(DefaultSocketPath)

	return &Daemon{
		SocketPath:  socketPath,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]struct{}),
	}
}

// Start initializes and runs the daemon
func (d *Daemon) Start() error {
	// Ensure socket doesn't already exist
	if err := os.RemoveAll(d.SocketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	var err error
	d.listener, err = net.Listen("unix", d.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	// Set appropriate permissions for the socket file
	if err := os.Chmod(d.SocketPath, 0660); err != nil {
		d.listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	fmt.Printf("Daemon started, listening on %s\n", d.SocketPath)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go d.acceptConnections()
	go d.handleSignals(sigChan)

	return nil
}

// acceptConnections accepts and handles incoming client connections
func (d *Daemon) acceptConnections() {
	for {
		select {
		case <-d.stopChan:
			return
		default:
			d.listener.(*net.UnixListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := d.listener.Accept()
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}
				fmt.Fprintf(os.Stderr, "Daemon accept error: %v\n", err)
				continue
			}

			// Track the connection
			d.connections[conn] = struct{}{}
			go d.handleConnection(conn)
		}
	}
}

// handleConnection processes a single client connection
func (d *Daemon) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		delete(d.connections, conn)
	}()

	// Create a buffer to read the command
	buffer := make([]byte, 128)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from client: %v\n", err)
		return
	}

	// Process the command
	command := string(buffer[:n])
	command = strings.TrimSpace(command)

	switch command {
	case "STOP":
		// Existing STOP command handler
		conn.Write([]byte("Stopping daemon\n"))
		go d.Stop()
	case "PING":
		// New PING command handler for status checks
		conn.Write([]byte("PONG\n"))
	default:
		// Handle other commands...
		fmt.Println("Received command:", command)
	}
}

// handleSignals processes system signals for graceful shutdown
func (d *Daemon) handleSignals(sigChan chan os.Signal) {
	<-sigChan
	fmt.Println("Received shutdown signal, stopping daemon...")
	d.Stop()
}

// Stop gracefully shuts down the daemon
func (d *Daemon) Stop() {
	// Signal the accept loop to stop
	close(d.stopChan)

	// Close listener
	if d.listener != nil {
		d.listener.Close()
	}

	// Close all active connections
	for conn := range d.connections {
		conn.Close()
	}

	// Clean up socket file
	os.RemoveAll(d.SocketPath)
	fmt.Println("Daemon stopped")
}

// Run starts the daemon and blocks until it's stopped
func Run(ctx context.Context) error {
	daemon := NewDaemon()

	if err := daemon.Start(); err != nil {
		return err
	}

	// Wait for context cancellation
	<-ctx.Done()
	daemon.Stop()
	return nil
}
