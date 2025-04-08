package daemon

import (
	"context"
	"errors"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/webserver"
	"io"
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
	processes   []ManagedProcess
	startMu     sync.Mutex
}

// NewDaemon creates a new Daemon instance with default configuration
func NewDaemon(socketPath string) *Daemon {
	return &Daemon{
		SocketPath:  socketPath,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]struct{}),
		processes:   make([]ManagedProcess, 0),
	}
}

// RegisterProcess adds a ManagedProcess to the daemon's list.
// Processes will be started in the order they are registered and
// stopped in the reverse order.
func (d *Daemon) RegisterProcess(p ManagedProcess) *Daemon {
	d.processes = append(d.processes, p)
	return d
}

// Start initializes and runs the daemon and all registered sub-processes.
func (d *Daemon) Start() error {
	d.startMu.Lock()
	defer d.startMu.Unlock()

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

	fmt.Printf("Daemon starting, listening on %s\n", d.SocketPath)

	startedProcesses := make([]ManagedProcess, 0, len(d.processes))
	for _, p := range d.processes {
		fmt.Printf("Starting process: %s...\n", p.Name())
		if err := p.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start process %s: %v\n", p.Name(), err)
			d.stopProcesses(startedProcesses, "startup rollback")
			d.listener.Close()
			os.RemoveAll(d.SocketPath)
			return fmt.Errorf("failed to start process %s: %w", p.Name(), err)
		}
		fmt.Printf("Process %s started successfully\n", p.Name())
		startedProcesses = append(startedProcesses, p)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.acceptConnections()
	}()

	go d.handleSignals(sigChan)

	fmt.Println("Daemon and all processes started successfully.")
	return nil
}

// Stop gracefully shuts down the daemon and all registered sub-processes.
func (d *Daemon) Stop() {
	d.startMu.Lock()
	defer d.startMu.Unlock()

	// Only run stop sequence once
	select {
	case <-d.stopChan:
		fmt.Println("Stop already in progress or completed.")
		return
	default:
		fmt.Println("Initiating daemon shutdown...")
		close(d.stopChan)
	}

	// First, stop accepting new connections
	if d.listener != nil {
		fmt.Println("Closing listener socket...")
		d.listener.Close()
	}

	// Close all active client connections (forcefully, consider graceful command handling if needed)
	fmt.Println("Closing active client connections...")
	for conn := range d.connections {
		conn.Close()
	}

	// Wait for acceptConnections goroutine to finish processing existing connections or timeout
	fmt.Println("Waiting for connection handler loop to exit...")
	d.wg.Wait()

	// --- Stop managed processes in reverse order ---
	fmt.Println("Stopping managed processes...")
	d.stopProcesses(d.processes, "shutdown")

	// Clean up socket file
	fmt.Println("Removing socket file...")
	os.RemoveAll(d.SocketPath)
	fmt.Println("Daemon stopped.")
}

// stopProcesses is a helper to stop a list of processes, typically in reverse order.
func (d *Daemon) stopProcesses(processesToStop []ManagedProcess, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var stopWg sync.WaitGroup
	errs := make(chan error, len(processesToStop))

	for i := len(processesToStop) - 1; i >= 0; i-- {
		p := processesToStop[i]
		stopWg.Add(1)
		go func(proc ManagedProcess) {
			defer stopWg.Done()
			fmt.Printf("Stopping process (%s): %s...\n", reason, proc.Name())
			if err := proc.Stop(ctx); err != nil {
				stopErr := fmt.Errorf("error stopping process %s: %w", proc.Name(), err)
				fmt.Fprintf(os.Stderr, "%v\n", stopErr)
				errs <- stopErr
			} else {
				fmt.Printf("Process %s stopped successfully\n", proc.Name())
			}
		}(p)
	}

	stopWg.Wait()
	close(errs)

	for err := range errs {
		fmt.Fprintf(os.Stderr, "Collected stop error (%s): %v\n", reason, err)
	}
	fmt.Printf("Finished stopping processes for reason: %s\n", reason)
}

func (d *Daemon) acceptConnections() {
	if d.listener == nil {
		fmt.Fprintf(os.Stderr, "Error: Listener not initialized in acceptConnections\n")
		return
	}
	fmt.Println("Starting connection accept loop...")

	for {
		select {
		case <-d.stopChan:
			fmt.Println("Accept loop received stop signal, exiting.")
			return
		default:
			// Set a deadline on the listener to prevent Accept() from blocking indefinitely.
			// This allows the loop to periodically check the stopChan.
			if unixListener, ok := d.listener.(*net.UnixListener); ok {
				err := unixListener.SetDeadline(time.Now().Add(1 * time.Second))
				if err != nil {
					// If setting deadline fails after listener is potentially closed, exit.
					if errors.Is(err, net.ErrClosed) {
						fmt.Println("Listener closed while setting deadline, exiting accept loop.")
						return
					}
					fmt.Fprintf(os.Stderr, "Error setting listener deadline: %v\n", err)

					// Avoid tight loop on persistent error
					time.Sleep(100 * time.Millisecond)
					continue
				}
			} else {
				fmt.Fprintf(os.Stderr, "Listener is not a *net.UnixListener, cannot set deadline. Exiting loop.\n")
				return
			}

			// Accept the next connection
			conn, err := d.listener.Accept()
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue
				}

				if errors.Is(err, net.ErrClosed) {
					fmt.Println("Listener closed, exiting accept loop.")
					return
				}

				fmt.Fprintf(os.Stderr, "Daemon accept error: %v\n", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Handle the connection in a new goroutine
			fmt.Println("Accepted new client connection from:", conn.RemoteAddr()) // RemoteAddr is nil for Unix sockets
			d.connections[conn] = struct{}{}
			go d.handleConnection(conn)
		}
	}
}

// handleConnection processes commands received from a single client connection.
func (d *Daemon) handleConnection(conn net.Conn) {
	defer func() {
		remoteAddr := conn.RemoteAddr() // Get address before closing
		fmt.Printf("Closing connection (%v)\n", remoteAddr)
		conn.Close()
		// TODO: Add mutex locking/unlocking if d.connections requires it for safe concurrent access.
		// d.connMutex.Lock()
		delete(d.connections, conn)
		// d.connMutex.Unlock()
		fmt.Printf("Connection closed and removed (%v)\n", remoteAddr)
	}()

	fmt.Printf("Handling connection (%v)...\n", conn.RemoteAddr())

	err := conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting read deadline for conn (%v): %v\n", conn.RemoteAddr(), err)
		return
	}

	buffer := make([]byte, 256)
	n, err := conn.Read(buffer)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			fmt.Fprintf(os.Stderr, "Client connection read timeout (%v)\n", conn.RemoteAddr())
			conn.Write([]byte("TIMEOUT: No command received.\n"))
			return
		}
		// io.EOF means the client closed the connection.
		if errors.Is(err, io.EOF) {
			fmt.Printf("Client (%v) closed connection (EOF).\n", conn.RemoteAddr())
			return
		}

		if errors.Is(err, net.ErrClosed) {
			fmt.Printf("Connection (%v) was closed before read completed.\n", conn.RemoteAddr())
			return
		}

		fmt.Fprintf(os.Stderr, "Error reading from client (%v): %v\n", conn.RemoteAddr(), err)
		return
	}

	// Reset the read deadline after successful read.
	conn.SetReadDeadline(time.Time{}) // No deadline

	// Process the received command.
	command := string(buffer[:n])
	command = strings.TrimSpace(command)

	fmt.Printf("Received command from (%v): '%s'\n", conn.RemoteAddr(), command)

	var response string

	err = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting write deadline for conn (%v): %v\n", conn.RemoteAddr(), err)
		return
	}

	// Handle recognized commands.
	switch command {
	case "STOP":
		response = "OK: Daemon stop acknowledged. Initiating shutdown...\n"
		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing STOP response to client (%v): %v\n", conn.RemoteAddr(), writeErr)
		}
		fmt.Println("STOP command received, triggering daemon shutdown asynchronously.")
		go d.Stop()

	case "PING":
		response = "PONG\n"
		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing PING response to client (%v): %v\n", conn.RemoteAddr(), writeErr)
		}

	case "STATUS":
		connCount := len(d.connections)
		processCount := len(d.processes)
		response = fmt.Sprintf("OK: Status: %d active client connection(s), %d managed process(es).\n", connCount, processCount)
		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing STATUS response to client (%v): %v\n", conn.RemoteAddr(), writeErr)
		}

	default:
		response = fmt.Sprintf("ERROR: Unknown command '%s'. Available commands: PING, STOP, STATUS\n", command)
		_, writeErr := conn.Write([]byte(response))
		if writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing error response to client (%v): %v\n", conn.RemoteAddr(), writeErr)
		}
	}

	conn.SetWriteDeadline(time.Time{})

	fmt.Printf("Finished handling command '%s' for connection (%v)\n", command, conn.RemoteAddr())
}

// handleSignals processes system signals for graceful shutdown
func (d *Daemon) handleSignals(sigChan chan os.Signal) {
	sig := <-sigChan
	fmt.Printf("Received signal %v, stopping daemon...\n", sig)
	d.Stop()
}
