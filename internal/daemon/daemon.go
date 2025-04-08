package daemon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode" // For sanitizing example
)

// CommandFunc defines the function signature for command handlers.
// Implementations MUST validate arguments and respect the provided context deadline/cancellation.
type CommandFunc func(ctx context.Context, args []string) (response string, err error)

// Config holds the configuration for the daemon
type Config struct {
	SocketPath         string
	ShutdownTimeout    time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	CommandExecTimeout time.Duration
	Logger             *slog.Logger
	MaxConnections     int // Max concurrent client connections (0 for unlimited)
}

// Daemon represents the main daemon structure
type Daemon struct {
	config      Config
	listener    net.Listener
	stopOnce    sync.Once
	stopChan    chan struct{}
	wg          sync.WaitGroup
	connections map[net.Conn]struct{}
	connMu      sync.RWMutex
	commands    map[string]CommandFunc
	cmdMu       sync.RWMutex
	logger      *slog.Logger
}

const defaultReaderSize = 4096 // Max command line length

// NewDaemon creates a new Daemon instance with the provided configuration
func NewDaemon(cfg Config) *Daemon {
	// --- Set Defaults ---
	if cfg.Logger == nil {
		// Default to stderr logger if none provided
		cfg.Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = 30 * time.Second
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 10 * time.Second
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 10 * time.Second
	}
	if cfg.CommandExecTimeout == 0 {
		cfg.CommandExecTimeout = 5 * time.Second
	}
	// cfg.MaxConnections defaults to 0 (unlimited) if not set

	d := &Daemon{
		config:      cfg,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]struct{}),
		commands:    make(map[string]CommandFunc),
		logger:      cfg.Logger,
	}

	d.registerDefaultCommands()
	return d
}

// registerDefaultCommands sets up the built-in commands
func (d *Daemon) registerDefaultCommands() {
	d.RegisterCommand("PING", d.handlePing)
	d.RegisterCommand("STATUS", d.handleStatus)
	d.RegisterCommand("STOP", d.handleStop)
}

// RegisterCommand adds or replaces a command handler. Not safe for concurrent use after Start().
func (d *Daemon) RegisterCommand(name string, handler CommandFunc) {
	d.cmdMu.Lock()
	defer d.cmdMu.Unlock()
	upperName := strings.ToUpper(name)
	if _, exists := d.commands[upperName]; exists {
		d.logger.Warn("Overwriting existing command handler", "command", upperName)
	}
	d.commands[upperName] = handler
	d.logger.Debug("Registered command", "command", upperName)
}

// Start initializes the listener and begins accepting connections.
func (d *Daemon) Start() error {
	select {
	case <-d.stopChan:
		return errors.New("daemon is stopped or stopping")
	default:
	}

	// --- Umask Handling (Feedback #1) ---
	// Set umask to ensure predictable socket permissions, restore previous umask on exit
	oldMask := syscall.Umask(0o002) // Allows 0775, results in 0660 after Chmod
	d.logger.Debug("Set umask", "new_mask", fmt.Sprintf("%04o", 0o002), "old_mask", fmt.Sprintf("%04o", oldMask))
	defer func() {
		syscall.Umask(oldMask)
		d.logger.Debug("Restored umask", "old_mask", fmt.Sprintf("%04o", oldMask))
	}()
	// --- End Umask Handling ---

	if err := os.RemoveAll(d.config.SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		d.logger.Error("Failed to remove existing socket file", "path", d.config.SocketPath, "error", err)
		return fmt.Errorf("failed to remove existing socket %s: %w", d.config.SocketPath, err)
	}

	var err error
	d.listener, err = net.Listen("unix", d.config.SocketPath)
	if err != nil {
		d.logger.Error("Failed to listen on socket", "path", d.config.SocketPath, "error", err)
		return fmt.Errorf("failed to listen on socket %s: %w", d.config.SocketPath, err)
	}

	cleanupListener := true
	defer func() {
		if cleanupListener && d.listener != nil {
			d.listener.Close()
			os.RemoveAll(d.config.SocketPath) // Attempt cleanup if Start fails mid-way
		}
	}()

	// Set permissions explicitly AFTER creating the socket
	if err = os.Chmod(d.config.SocketPath, 0660); err != nil {
		d.logger.Error("Failed to set socket permissions", "path", d.config.SocketPath, "permissions", "0660", "error", err)
		return fmt.Errorf("failed to set socket permissions for %s: %w", d.config.SocketPath, err)
	}
	d.logger.Info("Socket created", "path", d.config.SocketPath, "permissions", "0660")

	d.logger.Info("Daemon starting listener loop", "socket", d.config.SocketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go d.handleSignals(sigChan) // Handles OS signals

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.acceptConnections() // Handles incoming connections
	}()

	cleanupListener = false // Listener ownership transferred to acceptConnections goroutine
	d.logger.Info("Daemon started successfully")
	return nil
}

// Stop initiates graceful shutdown of the daemon.
func (d *Daemon) Stop() {
	d.stopOnce.Do(func() {
		d.logger.Info("Initiating daemon shutdown...")
		close(d.stopChan) // Signal all loops to stop

		// Close the listener first to stop accepting new connections
		if d.listener != nil {
			d.logger.Info("Closing listener socket", "path", d.config.SocketPath)
			if err := d.listener.Close(); err != nil {
				// Log non-standard errors during close
				if !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "use of closed network connection") {
					d.logger.Error("Error closing listener", "path", d.config.SocketPath, "error", err)
				}
			}
		}

		// Close existing client connections
		d.closeConnections()

		// Wait for the accept loop and all connection handlers to finish, with timeout
		d.logger.Info("Waiting for active connections and loops to finish...", "timeout", d.config.ShutdownTimeout)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), d.config.ShutdownTimeout)
		defer cancel()

		done := make(chan struct{})
		go func() {
			d.wg.Wait() // Wait for wg counter to reach zero
			close(done)
		}()

		select {
		case <-done:
			d.logger.Info("All connections and loops finished gracefully.")
		case <-shutdownCtx.Done():
			d.logger.Warn("Shutdown timeout exceeded waiting for active connections/loops.")
		}

		// Clean up the socket file
		d.logger.Info("Removing socket file", "path", d.config.SocketPath)
		if err := os.RemoveAll(d.config.SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			d.logger.Error("Failed to remove socket file during shutdown", "path", d.config.SocketPath, "error", err)
		}

		d.logger.Info("Daemon stopped.")
		// Consider adding application-specific cleanup hooks here if needed
	})
}

// closeConnections closes all tracked active client connections.
func (d *Daemon) closeConnections() {
	d.connMu.Lock()
	// Create a slice of connections to close outside the lock
	connsToClose := make([]net.Conn, 0, len(d.connections))
	for conn := range d.connections {
		connsToClose = append(connsToClose, conn)
	}
	// Clear the map while holding the lock
	d.connections = make(map[net.Conn]struct{})
	connCount := len(connsToClose)
	d.connMu.Unlock() // Release lock before closing connections

	if connCount == 0 {
		d.logger.Debug("No active client connections to close.")
		return
	}

	d.logger.Info("Closing active client connections", "count", connCount)

	// Close connections concurrently
	var closeWg sync.WaitGroup
	closeWg.Add(connCount)
	for _, conn := range connsToClose {
		go func(c net.Conn) {
			defer closeWg.Done()
			// Optionally set a very short deadline for a final message, but usually just close
			// c.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
			// _, _ = c.Write([]byte("SERVER_SHUTTING_DOWN\n"))
			c.Close()
		}(conn)
	}
	closeWg.Wait() // Wait for all Close calls to complete
	d.logger.Debug("Finished closing client connections.")
}

// acceptConnections runs the loop to accept incoming connections.
func (d *Daemon) acceptConnections() {
	d.logger.Info("Starting connection accept loop")

	for {
		// Set a deadline on the accept call to allow periodic checks of stopChan
		if unixListener, ok := d.listener.(*net.UnixListener); ok {
			if err := unixListener.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
				if errors.Is(err, net.ErrClosed) {
					d.logger.Info("Listener closed (detected in SetDeadline), exiting accept loop.")
					return // Exit loop cleanly
				}
				// Log unexpected errors setting deadline, pause briefly
				d.logger.Error("Error setting listener deadline", "error", err)
				time.Sleep(100 * time.Millisecond) // Avoid tight loop on persistent error
				// continue or return might be appropriate depending on error severity
			}
		} // Non-unix listeners might block longer

		// Attempt to accept a connection
		conn, err := d.listener.Accept()
		if err != nil {
			// Check if the loop should stop *first* before interpreting the error
			select {
			case <-d.stopChan:
				d.logger.Info("Stop signal received, exiting accept loop.")
				return
			default:
				// Continue checking the error if not stopping
			}

			// Handle common non-fatal errors
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue // Expected error due to deadline, loop again
			}
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				d.logger.Info("Listener closed while accepting, exiting accept loop.")
				return // Exit loop cleanly on expected closure
			}

			// Log other unexpected accept errors and pause briefly
			d.logger.Error("Daemon accept error", "error", err)
			time.Sleep(100 * time.Millisecond) // Avoid busy-looping
			continue
		}

		// Check if daemon is stopping *after* accepting but *before* handling
		select {
		case <-d.stopChan:
			d.logger.Info("Daemon stopping, rejecting newly accepted connection.", "remote_addr", conn.RemoteAddr())
			conn.Close()
			continue // Go back to check stopChan again explicitly
		default:
			// Not stopping, proceed to handle connection
		}

		// --- Connection Limiting (Additional #1) ---
		remoteAddr := conn.RemoteAddr() // Get once
		d.connMu.Lock()
		if d.config.MaxConnections > 0 && len(d.connections) >= d.config.MaxConnections {
			d.connMu.Unlock() // Unlock before logging/closing
			d.logger.Warn("Max connections reached, rejecting new connection", "limit", d.config.MaxConnections, "remote_addr", remoteAddr)
			// Optionally write a "BUSY" message before closing
			conn.Close()
			continue
		}

		// Add connection to map and increment WaitGroup *before* starting handler
		d.connections[conn] = struct{}{}
		d.wg.Add(1) // Increment counter for the connection handler goroutine
		currentConns := len(d.connections)
		d.connMu.Unlock()
		// --- End Connection Limiting ---

		d.logger.Info("Accepted new client connection", "remote_addr", remoteAddr, "current_connections", currentConns)

		// Launch handler in a goroutine
		go func(c net.Conn) {
			// Decrement counter when handler exits, regardless of reason
			defer d.wg.Done()
			d.handleConnection(c)
		}(conn)
	}
}

// sanitize for logging - replace non-printable chars (example implementation)
func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' { // Allow common whitespace too
			return r
		}
		return '?' // Replace non-printable with '?'
	}, s)
}

// handleConnection reads commands, executes them, and writes responses for a single connection.
func (d *Daemon) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	// Defer ensures cleanup even on panics within the handler (though panics should be avoided)
	defer func() {
		conn.Close() // Ensure connection is closed
		d.connMu.Lock()
		delete(d.connections, conn) // Remove from tracking map
		connCount := len(d.connections)
		d.connMu.Unlock()
		d.logger.Info("Connection closed and removed", "remote_addr", remoteAddr, "remaining_connections", connCount)
	}()

	d.logger.Debug("Handling connection", "remote_addr", remoteAddr)
	// --- Limit Reader Size (Feedback #6) ---
	reader := bufio.NewReaderSize(conn, defaultReaderSize)

	// Optional: Implement heartbeats here if needed (Additional #2)

	for {
		// Check for daemon stop signal before attempting to read
		select {
		case <-d.stopChan:
			d.logger.Info("Stop signal received during handling, closing connection", "remote_addr", remoteAddr)
			return
		default:
			// Continue processing commands
		}

		// Set read deadline for receiving the next command
		if d.config.ReadTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(d.config.ReadTimeout)); err != nil {
				// Log error but potentially continue if deadline setting fails transiently?
				// More robust might be to return here.
				d.logger.Error("Failed to set read deadline", "remote_addr", remoteAddr, "error", err)
				return // Close connection if deadline can't be set
			}
		}

		// Read command line (until newline)
		commandLine, err := reader.ReadString('\n')

		// --- IMPORTANT: Clear deadline immediately after read attempt ---
		// Prevents subsequent operations (like writing) from using the old read deadline
		if d.config.ReadTimeout > 0 {
			// Ignore error on clearing deadline, connection might already be closed/failed
			_ = conn.SetReadDeadline(time.Time{})
		}
		// --- End Clear Deadline ---

		// Handle read errors
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				d.logger.Warn("Client connection read timeout", "remote_addr", remoteAddr, "timeout", d.config.ReadTimeout)
				// Optionally try to inform client, ignore error as conn is likely dead
				// _ = d.writeResponse(conn, "TIMEOUT: No command received.\n", remoteAddr)
				return // Close connection on timeout
			}
			if errors.Is(err, io.EOF) {
				d.logger.Info("Client closed connection (EOF)", "remote_addr", remoteAddr)
				return // Normal closure by client
			}
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				d.logger.Info("Connection closed while reading", "remote_addr", remoteAddr)
				return // Connection closed, likely during shutdown or by client abruptly
			}
			if errors.Is(err, bufio.ErrBufferFull) {
				d.logger.Error("Command line exceeded buffer size", "remote_addr", remoteAddr, "limit", defaultReaderSize)
				_ = d.writeResponse(conn, "ERROR: Command too long.\n", remoteAddr)
				// Consider returning here or trying to read again if protocol allows recovery
				return
			}
			// Log other unexpected errors
			d.logger.Error("Error reading from client", "remote_addr", remoteAddr, "error", err)
			return // Close connection on unexpected error
		}

		// Process the received command line
		trimmedCmd := strings.TrimSpace(commandLine)
		if trimmedCmd == "" {
			continue // Ignore empty lines
		}

		// --- Log Sanitized Command (Feedback #5) ---
		d.logger.Debug("Received command line", "remote_addr", remoteAddr, "command_line", sanitize(trimmedCmd))

		// Parse command and arguments
		parts := strings.Fields(trimmedCmd)
		commandName := strings.ToUpper(parts[0])
		args := parts[1:] // Command handlers MUST validate these args

		var response string
		var cmdErr error

		// Look up command handler
		d.cmdMu.RLock()
		handler, found := d.commands[commandName]
		d.cmdMu.RUnlock()

		if found {
			// Execute command with timeout
			cmdCtx, cmdCancel := context.WithTimeout(context.Background(), d.config.CommandExecTimeout)
			response, cmdErr = handler(cmdCtx, args) // Handler must respect cmdCtx!
			cmdCancel()                              // Release context resources promptly

			if errors.Is(cmdErr, context.DeadlineExceeded) {
				d.logger.Error("Command execution timed out", "remote_addr", remoteAddr, "command", commandName, "timeout", d.config.CommandExecTimeout)
				// Overwrite cmdErr for clearer client message
				cmdErr = fmt.Errorf("command '%s' timed out after %v", commandName, d.config.CommandExecTimeout)
			}
		} else {
			cmdErr = fmt.Errorf("unknown command '%s'", commandName)
		}

		// Prepare and send response
		if cmdErr != nil {
			// Log command execution errors
			d.logger.Error("Command execution failed", "remote_addr", remoteAddr, "command", commandName, "args", args, "error", cmdErr)
			response = fmt.Sprintf("ERROR: %v\n", cmdErr)
		} else {
			// Ensure response has a newline trailer
			if !strings.HasSuffix(response, "\n") {
				response += "\n"
			}
			// Add "OK: " prefix convention for successful commands (except PING)
			if commandName != "PING" && !strings.HasPrefix(response, "OK:") && !strings.HasPrefix(response, "ERROR:") {
				response = "OK: " + response
			}
			d.logger.Debug("Command execution successful", "remote_addr", remoteAddr, "command", commandName) // Don't log response content by default
		}

		// Write response back to client
		writeErr := d.writeResponse(conn, response, remoteAddr)
		if writeErr != nil {
			// Error logged by writeResponse, terminate handler for this connection
			return
		}

		// Special handling for STOP command: exit handler loop after sending response
		// The main Stop() function will handle closing this connection eventually.
		if commandName == "STOP" && cmdErr == nil {
			d.logger.Info("STOP command processed successfully by handler, connection handler exiting.", "remote_addr", remoteAddr)
			return
		}

		// Continue loop to handle next command from this client
	}
}

// writeResponse sends a response string to the client with a write deadline.
func (d *Daemon) writeResponse(conn net.Conn, response string, remoteAddr string) error {
	// Set write deadline for this specific write operation
	if d.config.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(d.config.WriteTimeout)); err != nil {
			d.logger.Error("Failed to set write deadline", "remote_addr", remoteAddr, "error", err)
			// Continue attempt to write, but it might fail quickly
		}
		// Defer clearing the deadline after the write attempt completes
		defer func() {
			// Ignore error on clearing deadline
			_ = conn.SetWriteDeadline(time.Time{})
		}()
	}

	// Perform the write (Note: No explicit partial write handling loop here)
	n, writeErr := conn.Write([]byte(response)) // Using conn.Write directly

	// Log write outcome
	logArgs := []any{"remote_addr", remoteAddr, "response_len", len(response), "bytes_written", n}
	if writeErr != nil {
		if netErr, ok := writeErr.(net.Error); ok && netErr.Timeout() {
			d.logger.Error("Timeout writing response to client", append(logArgs, "error", writeErr)...)
		} else if errors.Is(writeErr, net.ErrClosed) || strings.Contains(writeErr.Error(), "use of closed network connection") {
			d.logger.Warn("Failed to write response, connection closed", append(logArgs, "error", writeErr)...)
		} else {
			// Includes potential partial write indication if n < len(response) without error
			d.logger.Error("Error writing response to client", append(logArgs, "error", writeErr)...)
		}
		return writeErr // Return the error
	}

	// Check for unexpected partial write (n < len(response) without error) - less common on UDS
	if n < len(response) {
		d.logger.Warn("Partial write occurred", logArgs...)
		// Potentially return an error or attempt retry, though simple return is often okay
		return io.ErrShortWrite
	}

	d.logger.Debug("Successfully wrote response", logArgs...)
	return nil
}

// handleSignals catches OS signals for shutdown.
func (d *Daemon) handleSignals(sigChan chan os.Signal) {
	sig := <-sigChan // Wait for signal
	d.logger.Info("Received OS signal, initiating shutdown...", "signal", sig)
	signal.Stop(sigChan) // Stop listening for more signals on this channel
	// Don't close sigChan here if other parts might still select on it (though unlikely here)
	go d.Stop() // Initiate graceful shutdown asynchronously
}

// --- Default Command Handlers ---

// handlePing responds PONG to a PING command.
func (d *Daemon) handlePing(ctx context.Context, args []string) (string, error) {
	// Example of context check (though PING is too fast to timeout usually)
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("ping cancelled: %w", ctx.Err())
	default:
		return "PONG", nil
	}
}

// handleStatus provides basic daemon status.
func (d *Daemon) handleStatus(ctx context.Context, args []string) (string, error) {
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

	sort.Strings(cmdNames) // Consistent order

	// Example of context check in a potentially longer operation
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("status cancelled: %w", ctx.Err())
	default:
		// Format status string
		status := fmt.Sprintf(
			"Connections: %d active (Limit: %d)\nCommands: %d registered (%s)",
			connCount,
			d.config.MaxConnections, // Show limit if set
			cmdCount,
			strings.Join(cmdNames, ", "),
		)
		// Consider adding uptime, memory usage, etc. here later
		return status, nil
	}
}

// handleStop triggers the daemon shutdown asynchronously.
// Note: Response delivery isn't guaranteed due to potential immediate shutdown race.
func (d *Daemon) handleStop(ctx context.Context, args []string) (string, error) {
	d.logger.Info("STOP command received via connection, triggering daemon shutdown.")
	// Trigger stop in a separate goroutine to allow response write attempt.
	go d.Stop()
	// The connection handler will exit after this response is written (or write fails)
	return "Daemon stop initiated.", nil
}
