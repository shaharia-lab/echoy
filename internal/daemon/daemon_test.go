package daemon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewDaemon(t *testing.T) {
	t.Parallel()

	defaultShutdownTimeout := 30 * time.Second
	defaultReadTimeout := 10 * time.Second
	defaultWriteTimeout := 10 * time.Second
	defaultCommandExecTimeout := 5 * time.Second
	defaultMaxConnections := 0

	testCases := []struct {
		name         string
		inputCfg     Config
		expectedCfg  Config
		expectLogger bool
	}{
		{
			name: "Zero Config",
			inputCfg: Config{
				SocketPath: "/tmp/zero.sock",
			},
			expectedCfg: Config{
				SocketPath:         "/tmp/zero.sock",
				ShutdownTimeout:    defaultShutdownTimeout,
				ReadTimeout:        defaultReadTimeout,
				WriteTimeout:       defaultWriteTimeout,
				CommandExecTimeout: defaultCommandExecTimeout,
				MaxConnections:     defaultMaxConnections,
			},
			expectLogger: true,
		},
		{
			name: "Partial Config With Logger",
			inputCfg: Config{
				SocketPath:      "/tmp/partial.sock",
				ReadTimeout:     5 * time.Second,
				MaxConnections:  50,
				Logger:          slog.New(slog.NewTextHandler(io.Discard, nil)),
				ShutdownTimeout: 15 * time.Second,
			},
			expectedCfg: Config{
				SocketPath:         "/tmp/partial.sock",
				ShutdownTimeout:    15 * time.Second,
				ReadTimeout:        5 * time.Second,
				WriteTimeout:       defaultWriteTimeout,
				CommandExecTimeout: defaultCommandExecTimeout,
				MaxConnections:     50,
			},
			expectLogger: true,
		},
		{
			name: "Full Config",
			inputCfg: Config{
				SocketPath:         "/tmp/full.sock",
				ShutdownTimeout:    60 * time.Second,
				ReadTimeout:        15 * time.Second,
				WriteTimeout:       15 * time.Second,
				CommandExecTimeout: 10 * time.Second,
				Logger:             slog.New(slog.NewJSONHandler(io.Discard, nil)),
				MaxConnections:     100,
			},
			expectedCfg: Config{
				SocketPath:         "/tmp/full.sock",
				ShutdownTimeout:    60 * time.Second,
				ReadTimeout:        15 * time.Second,
				WriteTimeout:       15 * time.Second,
				CommandExecTimeout: 10 * time.Second,
				MaxConnections:     100,
			},
			expectLogger: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			d := NewDaemon(tc.inputCfg)

			if d == nil {
				t.Fatal("NewDaemon returned nil")
			}

			if d.config.SocketPath != tc.expectedCfg.SocketPath {
				t.Errorf("expected SocketPath %q, got %q", tc.expectedCfg.SocketPath, d.config.SocketPath)
			}
			if d.config.ShutdownTimeout != tc.expectedCfg.ShutdownTimeout {
				t.Errorf("expected ShutdownTimeout %v, got %v", tc.expectedCfg.ShutdownTimeout, d.config.ShutdownTimeout)
			}
			if d.config.ReadTimeout != tc.expectedCfg.ReadTimeout {
				t.Errorf("expected ReadTimeout %v, got %v", tc.expectedCfg.ReadTimeout, d.config.ReadTimeout)
			}
			if d.config.WriteTimeout != tc.expectedCfg.WriteTimeout {
				t.Errorf("expected WriteTimeout %v, got %v", tc.expectedCfg.WriteTimeout, d.config.WriteTimeout)
			}
			if d.config.CommandExecTimeout != tc.expectedCfg.CommandExecTimeout {
				t.Errorf("expected CommandExecTimeout %v, got %v", tc.expectedCfg.CommandExecTimeout, d.config.CommandExecTimeout)
			}
			if d.config.MaxConnections != tc.expectedCfg.MaxConnections {
				t.Errorf("expected MaxConnections %d, got %d", tc.expectedCfg.MaxConnections, d.config.MaxConnections)
			}

			if tc.expectLogger && d.logger == nil {
				t.Error("expected Logger to be non-nil, but got nil")
			}
			if tc.inputCfg.Logger == nil && d.logger == nil {
				t.Error("expected default Logger to be created, but got nil")
			}
			if tc.inputCfg.Logger != nil && d.logger != tc.inputCfg.Logger {
				t.Error("provided Logger instance was not assigned")
			}

			if d.stopChan == nil {
				t.Error("stopChan was not initialized")
			}
			if d.connections == nil {
				t.Error("connections map was not initialized")
			}
			if d.commands == nil {
				t.Error("commands map was not initialized")
			}
		})
	}
}

func TestSanitize(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty String", "", ""},
		{"Printable ASCII", "Hello World 123!", "Hello World 123!"},
		{"With Newline and Tab", "Line1\nLine2\tEnd", "Line1\nLine2\tEnd"},
		{"With Null Byte", "Before\x00After", "Before?After"},
		{"With Bell Char", "Ring\aRing", "Ring?Ring"},
		{"Mixed Printable and Non-Printable", "Good\x01Bad\x02End", "Good?Bad?End"},
		{"Extended Unicode Printable", "你好世界 éàç", "你好世界 éàç"},
		{"Non-Printable Control Chars", "\x1b[31mRed\x1b[0m", "?[31mRed?[0m"}, // ANSI escape codes
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := sanitize(tc.input)
			if result != tc.expected {
				t.Errorf("sanitize(%q): expected %q, got %q", tc.input, tc.expected, result)
			}
		})
	}
}

func tempSocketPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "daemon-test-*.sock")
	if err != nil {
		t.Fatalf("Failed to create temp file for socket path: %v", err)
	}
	path := f.Name()
	f.Close()
	os.Remove(path)
	t.Logf("Using temporary socket path: %s", path)
	return path
}

func testLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// createTestDaemon provides a helper to setup a daemon for connection tests
func createTestDaemon(t *testing.T, cfg Config) (*Daemon, string) {
	t.Helper()
	if cfg.SocketPath == "" {
		cfg.SocketPath = tempSocketPath(t)
	}
	if cfg.Logger == nil {
		// Default to discard logger for most tests unless specific logs needed
		cfg.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = 1 * time.Second // Use shorter timeouts for tests
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = 1 * time.Second
	}
	if cfg.CommandExecTimeout == 0 {
		cfg.CommandExecTimeout = 500 * time.Millisecond
	}

	d := NewDaemon(cfg)
	if d == nil {
		t.Fatal("NewDaemon returned nil")
	}
	return d, cfg.SocketPath
}

func waitForWg(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Errorf("timed out waiting for WaitGroup after %v", timeout)
	}
}

func TestHandleConnection_Ping(t *testing.T) {
	t.Parallel()

	d, _ := createTestDaemon(t, Config{})
	d.RegisterCommand("PING", DefaultPingHandler)

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.handleConnection(serverConn)
	}()

	_, err := clientConn.Write([]byte("PING\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	responseBytes := make([]byte, 128)
	n, err := clientConn.Read(responseBytes)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	response := string(responseBytes[:n])
	expectedResponse := "PONG\n"
	if response != expectedResponse {
		t.Errorf("Expected response %q, got %q", expectedResponse, response)
	}

	clientConn.Close()
	waitForWg(t, &wg, 2*time.Second)
}

func TestHandleConnection_UnknownCommand(t *testing.T) {
	t.Parallel()

	d, _ := createTestDaemon(t, Config{})

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.handleConnection(serverConn)
	}()

	unknownCmd := "NOSUCHCOMMAND"
	_, err := clientConn.Write([]byte(unknownCmd + "\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	responseBytes := make([]byte, 128)
	n, err := clientConn.Read(responseBytes)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	response := string(responseBytes[:n])
	expectedPrefix := fmt.Sprintf("ERROR: unknown command '%s'", unknownCmd)
	if !strings.HasPrefix(response, expectedPrefix) {
		t.Errorf("Expected response prefix %q, got %q", expectedPrefix, response)
	}

	clientConn.Close()
	waitForWg(t, &wg, 2*time.Second)
}

func TestHandleConnection_ReadTimeout(t *testing.T) {
	t.Parallel()

	readTimeout := 50 * time.Millisecond
	d, _ := createTestDaemon(t, Config{
		ReadTimeout: readTimeout,
	})

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.handleConnection(serverConn)
	}()

	waitForWg(t, &wg, readTimeout*3)

	readBuf := make([]byte, 1)
	n, err := clientConn.Read(readBuf)

	if err == nil {
		t.Errorf("Expected error reading from clientConn after server timeout, but got %d bytes", n)
	} else if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) && !strings.Contains(err.Error(), "closed") {
		t.Errorf("Expected EOF or closed error reading from clientConn, got: %v", err)
	}
}

func TestHandleConnection_WriteTimeout(t *testing.T) {
	t.Parallel()

	writeTimeout := 50 * time.Millisecond
	readTimeout := writeTimeout * 2

	var logBuf bytes.Buffer
	testLogger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	d, _ := createTestDaemon(t, Config{
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		Logger:       testLogger,
	})

	echoResponse := "This is the response to echo\n"
	d.RegisterCommand("ECHO", func(ctx context.Context, args []string) (string, error) {
		return echoResponse, nil
	})

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.handleConnection(serverConn)
	}()

	_, err := clientConn.Write([]byte("ECHO\n"))
	if err != nil {
		t.Fatalf("Client write command failed: %v", err)
	}

	waitForWg(t, &wg, writeTimeout*3)

	clientConn.Close()

	logOutput := logBuf.String()
	expectedLogMsg := "Timeout writing response to client"
	if !strings.Contains(logOutput, expectedLogMsg) {
		t.Errorf("Expected log message %q not found in logs:\n%s", expectedLogMsg, logOutput)
	}
}

func TestHandleConnection_CommandExecTimeout(t *testing.T) {
	t.Parallel()

	execTimeout := 50 * time.Millisecond
	readWriteTimeout := execTimeout * 4

	d, _ := createTestDaemon(t, Config{
		CommandExecTimeout: execTimeout,
		ReadTimeout:        readWriteTimeout,
		WriteTimeout:       readWriteTimeout,
	})

	sleepDuration := execTimeout * 2

	d.RegisterCommand("SLOW", func(ctx context.Context, args []string) (string, error) {
		select {
		case <-time.After(sleepDuration):
			return "Finally finished sleeping", nil
		case <-ctx.Done():
			return "", fmt.Errorf("handler context cancelled: %w", ctx.Err())
		}
	})

	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		d.handleConnection(serverConn)
	}()

	_, err := clientConn.Write([]byte("SLOW\n"))
	if err != nil {
		t.Fatalf("Client write command failed: %v", err)
	}

	responseBytes := make([]byte, 256)
	n, err := clientConn.Read(responseBytes)
	if err != nil {
		t.Fatalf("Client read failed: %v", err)
	}

	response := string(responseBytes[:n])
	expectedErrorMsg := fmt.Sprintf("ERROR: command 'SLOW' timed out after %v", execTimeout)

	if !strings.HasPrefix(response, expectedErrorMsg) {
		t.Errorf("Expected response prefix %q, got %q", expectedErrorMsg, response)
	}

	clientConn.Close()
	waitForWg(t, &wg, readWriteTimeout)
}

func TestHandleConnection_MaxConnections(t *testing.T) {
	maxConns := 2
	socketPath := tempSocketPath(t)
	defer os.RemoveAll(socketPath)

	d, _ := createTestDaemon(t, Config{
		SocketPath:         socketPath,
		MaxConnections:     maxConns,
		CommandExecTimeout: 5 * time.Second,
		ReadTimeout:        5 * time.Second,
	})

	d.RegisterCommand("WAIT", func(ctx context.Context, args []string) (string, error) {
		<-ctx.Done()
		return "Waited", ctx.Err()
	})

	err := d.Start()
	if err != nil {
		t.Fatalf("Daemon Start failed: %v", err)
	}
	defer d.Stop()

	time.Sleep(50 * time.Millisecond)

	establishedConns := make([]net.Conn, 0, maxConns)
	defer func() {
		for _, conn := range establishedConns {
			conn.Close()
		}
	}()

	for i := 0; i < maxConns; i++ {
		conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
		if err != nil {
			t.Fatalf("Dial failed for connection %d: %v", i+1, err)
		}
		establishedConns = append(establishedConns, conn)

		_, err = conn.Write([]byte("WAIT\n"))
		if err != nil {
			conn.Close()
			t.Fatalf("Write WAIT command failed for connection %d: %v", i+1, err)
		}
	}

	time.Sleep(50 * time.Millisecond)

	d.connMu.RLock()
	currentCount := len(d.connections)
	d.connMu.RUnlock()
	if currentCount != maxConns {
		t.Fatalf("Expected %d connections tracked by daemon, found %d", maxConns, currentCount)
	} else {
		t.Logf("Verified %d connections are tracked.", currentCount)
	}

	excessConn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		t.Fatalf("Dial failed for excess connection: %v", err)
	}
	defer excessConn.Close()

	excessConn.SetReadDeadline(time.Now().Add(1 * time.Second))
	buf := make([]byte, 1)
	n, readErr := excessConn.Read(buf)

	if readErr == nil {
		t.Errorf("Expected error reading from rejected connection, but read %d bytes", n)
	} else if !errors.Is(readErr, io.EOF) && !strings.Contains(readErr.Error(), "closed") && !strings.Contains(readErr.Error(), "reset by peer") && !strings.Contains(readErr.Error(), "broken pipe") {
		t.Errorf("Expected EOF or closed/reset error reading from rejected connection, got: %v", readErr)
	} else {
		t.Logf("Received expected error from rejected connection: %v", readErr)
	}

	d.connMu.RLock()
	finalCount := len(d.connections)
	d.connMu.RUnlock()
	if finalCount > maxConns {
		t.Errorf("Daemon tracked more connections (%d) than the limit (%d)", finalCount, maxConns)
	}
}
