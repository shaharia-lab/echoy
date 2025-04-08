package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Daemon struct {
	SocketPath  string
	listener    net.Listener
	stopChan    chan struct{}
	connections map[net.Conn]struct{}
	connMu      sync.RWMutex
	wg          sync.WaitGroup
	processes   []ManagedProcess
	startMu     sync.Mutex
}

func NewDaemon(socketPath string) *Daemon {
	return &Daemon{
		SocketPath:  socketPath,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]struct{}),
		processes:   make([]ManagedProcess, 0),
	}
}

func (d *Daemon) RegisterProcess(p ManagedProcess) *Daemon {
	d.processes = append(d.processes, p)
	return d
}

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
	defer func() {
		if err != nil && d.listener != nil {
			d.listener.Close()
			os.RemoveAll(d.SocketPath)
		}
	}()

	if err = os.Chmod(d.SocketPath, 0660); err != nil {
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	fmt.Printf("Daemon starting, listening on %s\n", d.SocketPath)

	startedProcesses := make([]ManagedProcess, 0, len(d.processes))
	for _, p := range d.processes {
		fmt.Printf("Starting process: %s...\n", p.Name())
		if err = p.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start process %s: %v\n", p.Name(), err)
			d.stopProcesses(startedProcesses, "startup rollback")
			// Return error, deferred cleanup will handle listener/socket
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
	return nil // Clear err variable before returning nil
}

func (d *Daemon) Stop() {
	d.startMu.Lock()

	select {
	case <-d.stopChan:
		fmt.Println("Stop already in progress or completed.")
		d.startMu.Unlock()
		return
	default:
		fmt.Println("Initiating daemon shutdown...")
		close(d.stopChan)
	}
	d.startMu.Unlock()

	if d.listener != nil {
		fmt.Println("Closing listener socket...")
		d.listener.Close()
	}

	d.connMu.Lock()
	fmt.Println("Closing active client connections...")
	closedConnections := make([]net.Conn, 0, len(d.connections))
	for conn := range d.connections {
		closedConnections = append(closedConnections, conn)
	}
	d.connMu.Unlock() // Unlock before closing connections

	for _, conn := range closedConnections {
		conn.Close() // Close connections outside the lock
	}

	fmt.Println("Waiting for connection handler loop to exit...")
	d.wg.Wait()

	fmt.Println("Stopping managed processes...")
	d.startMu.Lock() // Lock only to safely copy the slice
	processesToStop := make([]ManagedProcess, len(d.processes))
	copy(processesToStop, d.processes)
	d.startMu.Unlock()
	d.stopProcesses(processesToStop, "shutdown")

	fmt.Println("Removing socket file...")
	os.RemoveAll(d.SocketPath)
	fmt.Println("Daemon stopped.")
}

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
			var listenerDeadlineSet bool
			if unixListener, ok := d.listener.(*net.UnixListener); ok {
				err := unixListener.SetDeadline(time.Now().Add(1 * time.Second))
				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						fmt.Println("Listener closed while setting deadline, exiting accept loop.")
						return
					}
					fmt.Fprintf(os.Stderr, "Error setting listener deadline: %v\n", err)
					time.Sleep(100 * time.Millisecond)
					continue
				}
				listenerDeadlineSet = true
			} else if d.listener.Addr().Network() != "unix" { // Corrected string comparison
				fmt.Fprintf(os.Stderr, "Warning: Listener is not a standard Unix listener, cannot set deadline reliably. Accept might block.\n")
			}

			conn, err := d.listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if errors.Is(err, net.ErrClosed) {
					if !listenerDeadlineSet {
						fmt.Println("Listener closed (detected in Accept), exiting accept loop.")
					} // If deadline was set, ErrClosed handled above or here indicates normal shutdown
					return
				}
				fmt.Fprintf(os.Stderr, "Daemon accept error: %v\n", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			fmt.Println("Accepted new client connection")
			d.connMu.Lock()
			// Check if daemon is stopping before adding connection
			select {
			case <-d.stopChan:
				d.connMu.Unlock()
				conn.Close() // Close newly accepted connection if stopping
				fmt.Println("Daemon stopping, rejected new connection.")
				continue // Go back to check stopChan again
			default:
				// Not stopping, add connection
				d.connections[conn] = struct{}{}
			}
			d.connMu.Unlock()
			go d.handleConnection(conn)
		}
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	connID := fmt.Sprintf("%v", conn.RemoteAddr()) // Simple ID

	defer func() {
		fmt.Printf("Closing connection (%s)\n", connID)
		conn.Close()
		d.connMu.Lock()
		delete(d.connections, conn)
		d.connMu.Unlock()
		fmt.Printf("Connection closed and removed (%s)\n", connID)
	}()

	fmt.Printf("Handling connection (%s)...\n", connID)

	readCtx, readCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer readCancel()
	if dl, ok := readCtx.Deadline(); ok {
		err := conn.SetReadDeadline(dl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting read deadline for conn (%s): %v\n", connID, err)
			return
		}
	}

	buffer := make([]byte, 256)
	n, err := conn.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Fprintf(os.Stderr, "Client connection read timeout (%s)\n", connID)
			// Try to inform client, ignore error as connection might be dead
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			conn.Write([]byte("TIMEOUT: No command received.\n"))
			conn.SetWriteDeadline(time.Time{})
			return
		}
		if errors.Is(err, io.EOF) {
			fmt.Printf("Client (%s) closed connection (EOF).\n", connID)
			return
		}
		if errors.Is(err, net.ErrClosed) {
			fmt.Printf("Connection (%s) was closed before read completed.\n", connID)
			return
		}
		fmt.Fprintf(os.Stderr, "Error reading from client (%s): %v\n", connID, err)
		return
	}
	// Reset read deadline immediately after read
	conn.SetReadDeadline(time.Time{})

	command := string(buffer[:n])
	command = strings.TrimSpace(command)

	fmt.Printf("Received command from (%s): '%s'\n", connID, command)

	var response string
	var writeErr error

	respCtx, respCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer respCancel()

	if dl, ok := respCtx.Deadline(); ok { // Corrected use of Deadline()
		err := conn.SetWriteDeadline(dl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting write deadline for conn (%s): %v\n", connID, err)
			// Continue; attempt write but it might fail
		}
	} else {
		// Should not happen with WithTimeout, but good practice
		fmt.Fprintf(os.Stderr, "Could not get deadline from response context for conn (%s)\n", connID)
	}

	switch command {
	case "STOP":
		response = "OK: Daemon stop acknowledged. Initiating shutdown...\n"
		_, writeErr = conn.Write([]byte(response))
		if writeErr == nil {
			fmt.Println("STOP command received, triggering daemon shutdown asynchronously.")
			go d.Stop() // Trigger async
		}

	case "PING":
		response = "PONG\n"
		_, writeErr = conn.Write([]byte(response))

	case "STATUS":
		d.connMu.RLock()
		connCount := len(d.connections)
		d.connMu.RUnlock()
		// Assuming RegisterProcess is only called before Start
		processCount := len(d.processes)
		response = fmt.Sprintf("OK: Status: %d active client connection(s), %d managed process(es).\n", connCount, processCount)
		_, writeErr = conn.Write([]byte(response))

	default:
		response = fmt.Sprintf("ERROR: Unknown command '%s'. Available commands: PING, STOP, STATUS\n", command)
		_, writeErr = conn.Write([]byte(response))
	}

	if writeErr != nil {
		select {
		case <-respCtx.Done(): // Check if context timed out
			fmt.Fprintf(os.Stderr, "Timeout writing response '%s' to client (%s): %v\n", strings.TrimSpace(response), connID, respCtx.Err())
		default:
			if !errors.Is(writeErr, net.ErrClosed) {
				fmt.Fprintf(os.Stderr, "Error writing response '%s' to client (%s): %v\n", strings.TrimSpace(response), connID, writeErr)
			}
		}
	}

	conn.SetWriteDeadline(time.Time{})

	fmt.Printf("Finished handling command '%s' for connection (%s)\n", command, connID)
}

func (d *Daemon) handleSignals(sigChan chan os.Signal) {
	sig := <-sigChan
	fmt.Printf("Received signal %v, stopping daemon...\n", sig)
	d.Stop()
}
