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
	"unicode" // For sanitizing example
)

// CommandFunc defines the function signature for command handlers.
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
		cfg.CommandExecTimeout = 5 * time.Second
	}
	if cfg.MaxConnections == 0 {
		cfg.MaxConnections = 0
	}

	d := &Daemon{
		config:      cfg,
		stopChan:    make(chan struct{}),
		connections: make(map[net.Conn]struct{}),
		commands:    make(map[string]CommandFunc),
		logger:      cfg.Logger,
	}

	return d
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

	oldMask := syscall.Umask(0o002)
	d.logger.Debug("Set umask", "new_mask", fmt.Sprintf("%04o", 0o002), "old_mask", fmt.Sprintf("%04o", oldMask))
	defer func() {
		syscall.Umask(oldMask)
		d.logger.Debug("Restored umask", "old_mask", fmt.Sprintf("%04o", oldMask))
	}()

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
		d.logger.Error("Failed to set socket permissions", "path", d.config.SocketPath, "permissions", "0660", "error", err)
		return fmt.Errorf("failed to set socket permissions for %s: %w", d.config.SocketPath, err)
	}
	d.logger.Info("Socket created", "path", d.config.SocketPath, "permissions", "0660")

	d.logger.Info("Daemon starting listener loop", "socket", d.config.SocketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go d.handleSignals(sigChan)

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.acceptConnections()
	}()

	cleanupListener = false
	d.logger.Info("Daemon started successfully")
	return nil
}

// Stop initiates graceful shutdown of the daemon.
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

		d.logger.Info("Waiting for active connections and loops to finish...", "timeout", d.config.ShutdownTimeout)
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
			d.logger.Warn("Shutdown timeout exceeded waiting for active connections/loops.")
		}

		d.logger.Info("Removing socket file", "path", d.config.SocketPath)
		if err := os.RemoveAll(d.config.SocketPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			d.logger.Error("Failed to remove socket file during shutdown", "path", d.config.SocketPath, "error", err)
		}

		d.logger.Info("Daemon stopped.")
	})
}

// closeConnections closes all tracked active client connections.
func (d *Daemon) closeConnections() {
	d.connMu.Lock()
	connsToClose := make([]net.Conn, 0, len(d.connections))
	for conn := range d.connections {
		connsToClose = append(connsToClose, conn)
	}

	d.connections = make(map[net.Conn]struct{})
	connCount := len(connsToClose)
	d.connMu.Unlock()

	if connCount == 0 {
		d.logger.Debug("No active client connections to close.")
		return
	}

	d.logger.Info("Closing active client connections", "count", connCount)

	var closeWg sync.WaitGroup
	closeWg.Add(connCount)
	for _, conn := range connsToClose {
		go func(c net.Conn) {
			defer closeWg.Done()
			c.Close()
		}(conn)
	}
	closeWg.Wait()
	d.logger.Debug("Finished closing client connections.")
}

func (d *Daemon) acceptConnections() {
	d.logger.Info("Starting connection accept loop")

	for {
		if unixListener, ok := d.listener.(*net.UnixListener); ok {
			if err := unixListener.SetDeadline(time.Now().Add(1 * time.Second)); err != nil {
				if errors.Is(err, net.ErrClosed) {
					d.logger.Info("Listener closed (detected in SetDeadline), exiting accept loop.")
					return
				}
				d.logger.Error("Error setting listener deadline", "error", err)
				time.Sleep(100 * time.Millisecond)
			}
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
			time.Sleep(100 * time.Millisecond)
			continue
		}

		select {
		case <-d.stopChan:
			d.logger.Info("Daemon stopping, rejecting newly accepted connection.", "remote_addr", conn.RemoteAddr())
			conn.Close()
			continue
		default:
		}

		remoteAddr := conn.RemoteAddr()
		d.connMu.Lock()
		if d.config.MaxConnections > 0 && len(d.connections) >= d.config.MaxConnections {
			d.connMu.Unlock()
			d.logger.Warn("Max connections reached, rejecting new connection", "limit", d.config.MaxConnections, "remote_addr", remoteAddr)
			conn.Close()
			continue
		}

		d.connections[conn] = struct{}{}
		d.wg.Add(1)
		currentConns := len(d.connections)
		d.connMu.Unlock()

		d.logger.Info("Accepted new client connection", "remote_addr", remoteAddr, "current_connections", currentConns)

		go func(c net.Conn) {
			defer d.wg.Done()
			d.handleConnection(c)
		}(conn)
	}
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\t' {
			return r
		}
		return '?'
	}, s)
}

func (d *Daemon) handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	defer func() {
		conn.Close()
		d.connMu.Lock()
		delete(d.connections, conn)
		connCount := len(d.connections)
		d.connMu.Unlock()
		d.logger.Info("Connection closed and removed", "remote_addr", remoteAddr, "remaining_connections", connCount)
	}()

	d.logger.Debug("Handling connection", "remote_addr", remoteAddr)
	reader := bufio.NewReaderSize(conn, defaultReaderSize)

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

		if d.config.ReadTimeout > 0 {
			_ = conn.SetReadDeadline(time.Time{})
		}

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				d.logger.Warn("Client connection read timeout", "remote_addr", remoteAddr, "timeout", d.config.ReadTimeout)
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
			if errors.Is(err, bufio.ErrBufferFull) {
				d.logger.Error("Command line exceeded buffer size", "remote_addr", remoteAddr, "limit", defaultReaderSize)
				_ = d.writeResponse(conn, "ERROR: Command too long.\n", remoteAddr)
				return
			}

			d.logger.Error("Error reading from client", "remote_addr", remoteAddr, "error", err)
			return
		}

		trimmedCmd := strings.TrimSpace(commandLine)
		if trimmedCmd == "" {
			continue
		}

		d.logger.Debug("Received command line", "remote_addr", remoteAddr, "command_line", sanitize(trimmedCmd))

		parts := strings.Fields(trimmedCmd)
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
				d.logger.Error("Command execution timed out", "remote_addr", remoteAddr, "command", commandName, "timeout", d.config.CommandExecTimeout)
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
			if commandName != "PING" && !strings.HasPrefix(response, "OK:") && !strings.HasPrefix(response, "ERROR:") {
				response = "OK: " + response
			}
			d.logger.Debug("Command execution successful", "remote_addr", remoteAddr, "command", commandName)
		}

		writeErr := d.writeResponse(conn, response, remoteAddr)
		if writeErr != nil {
			return
		}

		if commandName == "STOP" && cmdErr == nil {
			d.logger.Info("STOP command processed successfully by handler, connection handler exiting.", "remote_addr", remoteAddr)
			return
		}
	}
}

func (d *Daemon) writeResponse(conn net.Conn, response string, remoteAddr string) error {
	if d.config.WriteTimeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(d.config.WriteTimeout)); err != nil {
			d.logger.Error("Failed to set write deadline", "remote_addr", remoteAddr, "error", err)
		}
		defer func() {
			_ = conn.SetWriteDeadline(time.Time{})
		}()
	}

	n, writeErr := conn.Write([]byte(response))

	logArgs := []any{"remote_addr", remoteAddr, "response_len", len(response), "bytes_written", n}
	if writeErr != nil {
		if netErr, ok := writeErr.(net.Error); ok && netErr.Timeout() {
			d.logger.Error("Timeout writing response to client", append(logArgs, "error", writeErr)...)
		} else if errors.Is(writeErr, net.ErrClosed) || strings.Contains(writeErr.Error(), "use of closed network connection") {
			d.logger.Warn("Failed to write response, connection closed", append(logArgs, "error", writeErr)...)
		} else {
			d.logger.Error("Error writing response to client", append(logArgs, "error", writeErr)...)
		}
		return writeErr
	}

	if n < len(response) {
		d.logger.Warn("Partial write occurred", logArgs...)
		return io.ErrShortWrite
	}

	d.logger.Debug("Successfully wrote response", logArgs...)
	return nil
}

func (d *Daemon) handleSignals(sigChan chan os.Signal) {
	sig := <-sigChan
	d.logger.Info("Received OS signal, initiating shutdown...", "signal", sig)
	signal.Stop(sigChan)
	go d.Stop()
}
