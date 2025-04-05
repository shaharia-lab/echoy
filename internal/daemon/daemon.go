package daemon

import (
	"context"
	"errors"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/webserver"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Daemon represents a Unix socket daemon service with an optional web server
type Daemon struct {
	SocketPath  string
	webServer   *webserver.WebServer
	listener    net.Listener
	stopChan    chan struct{}
	connections map[net.Conn]struct{}
	wg          sync.WaitGroup
}

// NewDaemon creates a new Daemon instance with default configuration
func NewDaemon(socketPath string) *Daemon {
	return &Daemon{
		SocketPath:  socketPath,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]struct{}),
	}
}

// WithWebServer attaches a webserver to the daemon
func (d *Daemon) WithWebServer(ws *webserver.WebServer) *Daemon {
	d.webServer = ws
	return d
}

// Start initializes and runs the daemon
func (d *Daemon) Start() error {
	if err := os.RemoveAll(d.SocketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	var err error
	d.listener, err = net.Listen("unix", d.SocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	if err := os.Chmod(d.SocketPath, 0660); err != nil {
		d.listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	// Start webserver if configured
	if d.webServer != nil {
		if err := d.webServer.Start(); err != nil {
			d.listener.Close()
			return fmt.Errorf("failed to start web server: %w", err)
		}
		fmt.Println("Web server started")
	}

	fmt.Printf("Daemon started, listening on %s\n", d.SocketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.acceptConnections()
	}()

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
		conn.Write([]byte("Stopping daemon\n"))
		go d.Stop()
	case "PING":
		conn.Write([]byte("PONG\n"))
	default:
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
	// Only run stop sequence once
	select {
	case <-d.stopChan:
		return
	default:
		close(d.stopChan)
	}

	// First, stop accepting new connections
	if d.listener != nil {
		d.listener.Close()
	}

	// Close all active connections
	for conn := range d.connections {
		conn.Close()
	}

	// Wait for acceptConnections goroutine to finish
	d.wg.Wait()

	// Stop the web server if it's running
	if d.webServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := d.webServer.Stop(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping web server: %v\n", err)
		} else {
			fmt.Println("Web server stopped")
		}
	}

	// Clean up socket file
	os.RemoveAll(d.SocketPath)
	fmt.Println("Daemon stopped")
}

// Run starts the daemon and blocks until it's stopped
func Run(ctx context.Context, socketPath string, apiPort string) error {
	daemon := NewDaemon(socketPath)

	// Only configure webserver if API port is provided
	if apiPort != "" {
		ws := webserver.NewWebServer(apiPort)
		daemon.WithWebServer(ws)
	}

	if err := daemon.Start(); err != nil {
		return err
	}

	// Wait for context cancellation
	<-ctx.Done()
	daemon.Stop()
	return nil
}
