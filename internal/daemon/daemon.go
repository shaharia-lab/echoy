package daemon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/types"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Config holds the configuration for the daemon
type Config struct {
	SocketPath         string
	ShutdownTimeout    time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	CommandExecTimeout time.Duration
	Logger             logger.Logger
	MaxConnections     int
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
	commands    map[string]types.CommandFunc
	cmdMu       sync.RWMutex
	logger      logger.Logger
	cancelCtx   context.CancelFunc
}

const defaultReaderSize = 4096

// NewDaemon creates a new Daemon instance with the provided configuration
func NewDaemon(cfg Config, logger logger.Logger) *Daemon {
	cfg.Logger = logger
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
		commands:    make(map[string]types.CommandFunc),
		logger:      cfg.Logger,
	}

	return d
}

func (d *Daemon) SetCancelFunc(cancelFunc context.CancelFunc) {
	d.cancelCtx = cancelFunc
}

// RegisterCommand adds or replaces a command handler. Not safe for concurrent use after Start().
func (d *Daemon) RegisterCommand(name string, handler types.CommandFunc) {
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

	// Use platform-specific umask setting
	defer setSocketUmask(d)()

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

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.acceptConnections()
	}()

	cleanupListener = false
	d.logger.Info("Daemon started successfully, entering wait state.") // New Log Message
	d.wg.Wait()                                                        // The blocking call
	d.logger.Info("Daemon main loop(s) finished.")
	return nil
}

// Stop initiates graceful shutdown of the daemon.
func (d *Daemon) Stop() {
	d.stopOnce.Do(func() {
		d.logger.Info("Stop: Initiating daemon shutdown...")

		if d.cancelCtx != nil {
			d.logger.Debug("Stop: Calling main context cancel function.")
			d.cancelCtx()
		} else {
			d.logger.Warn("Stop: Main context cancel function is nil.")
		}

		d.logger.Debug("Stop: Closing stopChan...")
		close(d.stopChan) // Signal internal loops

		if d.listener != nil {
			d.logger.Info("Stop: Closing listener socket", "path", d.config.SocketPath)
			if err := d.listener.Close(); err != nil {
				// ... (existing error handling) ...
			}
		} else {
			d.logger.Warn("Stop: Listener was nil, cannot close.")
		}

		d.logger.Debug("Stop: Closing client connections...")
		d.closeConnections() // Ensure this doesn't hang

		d.logger.Info("Stop: Waiting for active connections/loops...", "timeout", d.config.ShutdownTimeout)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), d.config.ShutdownTimeout)
		defer cancel()

		done := make(chan struct{})
		waitReturned := make(chan struct{}) // Channel to signal wg.Wait returned
		go func() {
			d.logger.Debug("Stop: Starting wg.Wait() in goroutine...")
			d.wg.Wait()
			d.logger.Debug("Stop: wg.Wait() finished.")
			close(done)         // Signal success
			close(waitReturned) // Signal wait finished for logging below
		}()

		select {
		case <-done:
			d.logger.Info("Stop: All connections and loops finished gracefully.")
		case <-shutdownCtx.Done():
			d.logger.Warn("Stop: Shutdown timeout exceeded waiting for active connections/loops.")
			// Even on timeout, we MUST wait for the wg.Wait goroutine to finish
			// otherwise we might race with wg.Done() calls later? Or is wg internally safe?
			// Let's wait briefly for safety/logging. wg itself is safe.
			<-waitReturned
		}

		d.logger.Info("Stop: Removing socket file", "path", d.config.SocketPath) // Log BEFORE removal
		err := os.RemoveAll(d.config.SocketPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			d.logger.Error("Stop: Failed to remove socket file during shutdown", "path", d.config.SocketPath, "error", err)
			// Consider adding a panic here ONLY for debugging if needed:
			// panic(fmt.Sprintf("PANIC: Failed to remove socket file %s: %v", d.config.SocketPath, err))
		} else if err == nil {
			d.logger.Info("Stop: Successfully removed socket file.") // Log success
		} else {
			d.logger.Info("Stop: Socket file was already removed.") // Log not exist
		}

		d.logger.Info("Stop: Daemon stopped method finished.") // Log exit
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
				if writeErr := d.writeResponse(conn, "ERROR: Command too long.\n", remoteAddr); writeErr != nil {
					d.logger.Warn("Failed to write 'Command too long' error to client", "remote_addr", remoteAddr, "error", writeErr)
				}
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

	logFields := map[string]interface{}{
		"remote_addr":   remoteAddr,
		"response_len":  len(response),
		"bytes_written": n,
	}
	if writeErr != nil {
		if netErr, ok := writeErr.(net.Error); ok && netErr.Timeout() {
			d.logger.WithField(logger.ErrorKey, writeErr).WithFields(logFields).Error("Timeout writing response to client")
		} else if errors.Is(writeErr, net.ErrClosed) || strings.Contains(writeErr.Error(), "use of closed network connection") {
			d.logger.WithField(logger.ErrorKey, writeErr).WithFields(logFields).Warn("Failed to write response, connection closed")
		} else {
			d.logger.WithField(logger.ErrorKey, writeErr).WithFields(logFields).Error("Error writing response to client")
		}
		return writeErr
	}

	if n < len(response) {
		d.logger.WithFields(logFields).Warn("Partial write to client")

		return io.ErrShortWrite
	}

	d.logger.WithFields(logFields).Debug("Successfully wrote response")

	return nil
}
