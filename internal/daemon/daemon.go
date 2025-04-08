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
	"strings"
	"sync"
	"syscall"
	"time"
)

type CommandFunc func(ctx context.Context, args []string) (response string, err error)

type Config struct {
	SocketPath         string
	ShutdownTimeout    time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	CommandExecTimeout time.Duration // Added timeout for individual command execution
	Logger             *slog.Logger
}

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

func NewDaemon(cfg Config) *Daemon {
	if cfg.Logger == nil {
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
		cfg.CommandExecTimeout = 5 * time.Second // Default command execution timeout
	}

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

func (d *Daemon) registerDefaultCommands() {
	d.RegisterCommand("PING", d.handlePing)
	d.RegisterCommand("STATUS", d.handleStatus)
	d.RegisterCommand("STOP", d.handleStop)
}

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

func (d *Daemon) Start() error {
	select {
	case <-d.stopChan:
		return errors.New("daemon is stopped or stopping")
	default:
	}

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
			os.RemoveAll(d.config.SocketPath)
		}
	}()

	if err = os.Chmod(d.config.SocketPath, 0660); err != nil {
		d.logger.Error("Failed to set socket permissions", "path", d.config.SocketPath, "error", err)
		return fmt.Errorf("failed to set socket permissions for %s: %w", d.config.SocketPath, err)
	}

	d.logger.Info("Daemon starting, listening on socket", "path", d.config.SocketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go d.handleSignals(sigChan)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.acceptConnections()
	}()

	cleanupListener = false // Listener ownership transferred to acceptConnections goroutine
	d.logger.Info("Daemon started successfully")
	return nil
}

func (d *Daemon) Stop() {
	d.stopOnce.Do(func() {
		d.logger.Info("Initiating daemon shutdown...")
		close(d.stopChan)

		if d.listener != nil {
			d.logger.Info("Closing listener socket", "path", d.config.SocketPath)
			if err := d.listener.Close(); err != nil {
				if !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "use of closed network connection") {
					d.logger.Error("Error closing listener", "path", d.config.SocketPath, "error", err)
				}
			}
		}

		d.closeConnections()

		d.logger.Info("Waiting for active connections and loops to finish...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), d.config.ShutdownTimeout)
		defer cancel()

		done := make(chan struct{})
		go func() {
			d.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			d.logger.Info("All connections and loops finished gracefully.")
		case <-shutdownCtx.Done():
			d.logger.Warn("Shutdown timeout exceeded waiting for active connections.")
		}

		d.logger.Info("Removing socket file", "path", d.config.SocketPath)
		if err := os.RemoveAll(d.config.SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			d.logger.Error("Failed to remove socket file during shutdown", "path", d.config.SocketPath, "error", err)
		}

		d.logger.Info("Daemon stopped.")
	})
}

func (d *Daemon) closeConnections() {
	d.connMu.Lock()
	if len(d.connections) == 0 {
		d.connMu.Unlock()
		d.logger.Debug("No active client connections to close.")
		return
	}

	d.logger.Info("Closing active client connections", "count", len(d.connections))
	connsToClose := make([]net.Conn, 0, len(d.connections))
	for conn := range d.connections {
		connsToClose = append(connsToClose, conn)
	}
	d.connections = make(map[net.Conn]struct{})
	d.connMu.Unlock()

	var closeWg sync.WaitGroup
	for _, conn := range connsToClose {
		closeWg.Add(1)
		go func(c net.Conn) {
			defer closeWg.Done()
			// Set a short deadline for potentially sending a shutdown message
			c.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
			// Optional: Attempt to write a shutdown message, ignore errors
			// _, _ = c.Write([]byte("SERVER_SHUTTING_DOWN\n"))
			c.Close()
		}(conn)
	}
	closeWg.Wait()
	d.logger.Debug("Finished closing client connections.")
}

func (d *Daemon) acceptConnections() {
	d.logger.Info("Starting connection accept loop")

	for {
		var acceptDeadline time.Time
		if unixListener, ok := d.listener.(*net.UnixListener); ok {
			acceptDeadline = time.Now().Add(1 * time.Second)
			if err := unixListener.SetDeadline(acceptDeadline); err != nil {
				if errors.Is(err, net.ErrClosed) {
					d.logger.Info("Listener closed (detected in SetDeadline), exiting accept loop.")
					return
				}
				d.logger.Error("Error setting listener deadline", "error", err)
			}
		} else if d.listener != nil {
			d.logger.Warn("Listener is not a standard Unix listener, cannot set non-blocking deadline.")
		} else {
			d.logger.Error("Listener is nil in acceptConnections, exiting loop")
			return
		}

		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.stopChan:
				d.logger.Info("Stop signal received, exiting accept loop.")
				return
			default:
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				d.logger.Info("Listener closed while accepting, exiting accept loop.")
				return
			}

			d.logger.Error("Daemon accept error", "error", err)
			time.Sleep(100 * time.Millisecond) // Avoid busy-looping on persistent errors
			continue
		}

		select {
		case <-d.stopChan:
			d.logger.Info("Daemon stopping, rejecting newly accepted connection.", "remote_addr", conn.RemoteAddr())
			conn.Close()
			continue
		default:
		}

		d.logger.Info("Accepted new client connection", "remote_addr", conn.RemoteAddr())

		d.connMu.Lock()
		d.connections[conn] = struct{}{}
		d.connMu.Unlock()

		d.wg.Add(1)
		go func(c net.Conn) {
			defer d.wg.Done()
			d.handleConnection(c)
		}(conn)
	}
}

func (d *Daemon) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	defer func() {
		d.logger.Debug("Closing connection handler", "remote_addr", remoteAddr)
		conn.Close()
		d.connMu.Lock()
		delete(d.connections, conn)
		d.connMu.Unlock()
		d.logger.Info("Connection closed and removed", "remote_addr", remoteAddr)
	}()

	d.logger.Debug("Handling connection", "remote_addr", remoteAddr)
	reader := bufio.NewReader(conn)

	for {
		select {
		case <-d.stopChan:
			d.logger.Info("Stop signal received during handling, closing connection", "remote_addr", remoteAddr)
			return
		default:
		}

		if d.config.ReadTimeout > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(d.config.ReadTimeout)); err != nil {
				d.logger.Error("Failed to set read deadline", "remote_addr", remoteAddr, "error", err)
				return
			}
		}

		commandLine, err := reader.ReadString('\n')

		// Clear deadline immediately after read attempt completes (success or fail)
		if d.config.ReadTimeout > 0 {
			// Ignoring error on clearing deadline as the connection might already be failed/closed
			_ = conn.SetReadDeadline(time.Time{})
		}

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				d.logger.Warn("Client connection read timeout", "remote_addr", remoteAddr)
				_ = d.writeResponse(conn, "TIMEOUT: No command received within timeout.\n", remoteAddr)
				return
			}
			if errors.Is(err, io.EOF) {
				d.logger.Info("Client closed connection (EOF)", "remote_addr", remoteAddr)
				return
			}
			if errors.Is(err, net.ErrClosed) || strings.Contains(err.Error(), "use of closed network connection") {
				d.logger.Info("Connection closed while reading", "remote_addr", remoteAddr)
				return
			}
			d.logger.Error("Error reading from client", "remote_addr", remoteAddr, "error", err)
			return
		}

		commandLine = strings.TrimSpace(commandLine)
		if commandLine == "" {
			continue
		}

		d.logger.Debug("Received command line", "remote_addr", remoteAddr, "command_line", commandLine)

		parts := strings.Fields(commandLine)
		commandName := strings.ToUpper(parts[0])
		args := parts[1:]

		var response string
		var cmdErr error

		d.cmdMu.RLock()
		handler, found := d.commands[commandName]
		d.cmdMu.RUnlock()

		if found {
			cmdCtx, cmdCancel := context.WithTimeout(context.Background(), d.config.CommandExecTimeout)
			response, cmdErr = handler(cmdCtx, args)
			cmdCancel()

			if errors.Is(cmdErr, context.DeadlineExceeded) {
				d.logger.Error("Command execution timed out", "remote_addr", remoteAddr, "command", commandName)
				// Overwrite cmdErr to provide a clearer error message to client
				cmdErr = fmt.Errorf("command '%s' timed out after %v", commandName, d.config.CommandExecTimeout)
			}

		} else {
			cmdErr = fmt.Errorf("unknown command '%s'", commandName)
		}

		if cmdErr != nil {
			d.logger.Error("Command execution failed", "remote_addr", remoteAddr, "command", commandName, "args", args, "error", cmdErr)
			response = fmt.Sprintf("ERROR: %v\n", cmdErr)
		} else {
			if !strings.HasSuffix(response, "\n") {
				response += "\n"
			}
			// Add OK prefix for non-error responses, unless it's PING or already has a prefix structure
			// A more robust approach might involve command handlers returning a struct with status/payload
			if commandName != "PING" && !strings.HasPrefix(response, "OK:") && !strings.HasPrefix(response, "ERROR:") {
				response = "OK: " + response
			}
			d.logger.Debug("Command execution successful", "remote_addr", remoteAddr, "command", commandName, "args", args, "response", strings.TrimSpace(response))
		}

		writeErr := d.writeResponse(conn, response, remoteAddr)
		if writeErr != nil {
			return // Error logged by writeResponse, terminate handler
		}

		if commandName == "STOP" && cmdErr == nil {
			d.logger.Info("STOP command processed successfully, handler exiting.", "remote_addr", remoteAddr)
			return // Exit handler loop after successful STOP response
		}
	}
}

func (d *Daemon) writeResponse(conn net.Conn, response string, remoteAddr string) error {
	if d.config.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(d.config.WriteTimeout)); err != nil {
			d.logger.Error("Failed to set write deadline", "remote_addr", remoteAddr, "error", err)
		}
		defer func() {
			// Ignoring error on clearing deadline
			_ = conn.SetWriteDeadline(time.Time{})
		}()
	}

	_, writeErr := conn.Write([]byte(response))
	if writeErr != nil {
		logArgs := []any{"remote_addr", remoteAddr, "response_len", len(response)}
		if netErr, ok := writeErr.(net.Error); ok && netErr.Timeout() {
			d.logger.Error("Timeout writing response to client", append(logArgs, "error", writeErr)...)
		} else if errors.Is(writeErr, net.ErrClosed) || strings.Contains(writeErr.Error(), "use of closed network connection") {
			d.logger.Warn("Failed to write response, connection closed", append(logArgs, "error", writeErr)...)
		} else {
			d.logger.Error("Error writing response to client", append(logArgs, "error", writeErr)...)
		}
		return writeErr
	}

	d.logger.Debug("Successfully wrote response", "remote_addr", remoteAddr, "response_len", len(response))
	return nil
}

func (d *Daemon) handleSignals(sigChan chan os.Signal) {
	sig := <-sigChan
	d.logger.Info("Received OS signal", "signal", sig)
	signal.Stop(sigChan)
	close(sigChan)
	go d.Stop() // Trigger stop asynchronously to allow signal handler to return quickly
}

// --- Default Command Handlers ---

func (d *Daemon) handlePing(ctx context.Context, args []string) (string, error) {
	return "PONG", nil
}

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

	// Sort command names for consistent output? Optional.
	// sort.Strings(cmdNames)

	status := fmt.Sprintf(
		"Connections: %d active\nCommands: %d registered (%s)",
		connCount,
		cmdCount,
		strings.Join(cmdNames, ", "),
	)
	// Add more status info like uptime, memory usage if needed
	return status, nil
}

func (d *Daemon) handleStop(ctx context.Context, args []string) (string, error) {
	d.logger.Info("STOP command received via connection, triggering daemon shutdown.")
	go d.Stop()
	return "Daemon stop initiated.", nil
}
