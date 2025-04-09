package daemon

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// MockConnection implements net.Conn for testing
type MockConnection struct {
	WriteData     []byte
	WriteError    error
	ReadData      string
	ReadError     error
	CloseError    error
	DeadlineCalls int
	readIndex     int
}

func (m *MockConnection) Read(b []byte) (n int, err error) {
	if m.ReadError != nil {
		return 0, m.ReadError
	}

	if m.readIndex >= len(m.ReadData) {
		return 0, io.EOF
	}

	n = copy(b, m.ReadData[m.readIndex:])
	m.readIndex += n

	return n, nil
}

func (m *MockConnection) Write(b []byte) (n int, err error) {
	if m.WriteError != nil {
		return 0, m.WriteError
	}
	m.WriteData = b
	return len(b), nil
}

func (m *MockConnection) Close() error                       { return m.CloseError }
func (m *MockConnection) LocalAddr() net.Addr                { return &net.UnixAddr{Name: "client", Net: "unix"} }
func (m *MockConnection) RemoteAddr() net.Addr               { return &net.UnixAddr{Name: "server", Net: "unix"} }
func (m *MockConnection) SetDeadline(t time.Time) error      { m.DeadlineCalls++; return nil }
func (m *MockConnection) SetReadDeadline(t time.Time) error  { m.DeadlineCalls++; return nil }
func (m *MockConnection) SetWriteDeadline(t time.Time) error { m.DeadlineCalls++; return nil }

// MockConnectionProvider implements ConnectionProvider for testing
/*type MockConnectionProvider struct {
	Conn  net.Conn
	Error error
}

func (m *MockConnectionProvider) Connect(ctx context.Context) (net.Conn, error) {
	return m.Conn, m.Error
}*/

func TestDaemonClient_Execute(t *testing.T) {
	tests := []struct {
		name          string
		mockConn      *MockConnection
		providerError error
		cmd           string
		want          string
		wantErr       bool
	}{
		{
			name: "successful command",
			mockConn: &MockConnection{
				ReadData: "SUCCESS\n",
			},
			cmd:     "TEST",
			want:    "SUCCESS",
			wantErr: false,
		},
		{
			name:          "connection error",
			providerError: errors.New("connection refused"),
			cmd:           "TEST",
			want:          "",
			wantErr:       true,
		},
		{
			name: "write error",
			mockConn: &MockConnection{
				WriteError: errors.New("write error"),
			},
			cmd:     "TEST",
			want:    "",
			wantErr: true,
		},
		{
			name: "read error",
			mockConn: &MockConnection{
				ReadError: errors.New("read error"),
			},
			cmd:     "TEST",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockConnectionProvider{
				Conn:  tt.mockConn,
				Error: tt.providerError,
			}

			client := &DaemonClient{
				Provider: provider,
			}

			got, err := client.Execute(context.Background(), tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("DaemonClient.Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DaemonClient.Execute() = %v, want %v", got, tt.want)
			}

			// Check that command was written correctly if no errors
			if tt.mockConn != nil && tt.providerError == nil && tt.mockConn.WriteError == nil {
				expectedCmd := tt.cmd
				if !strings.HasSuffix(expectedCmd, "\n") {
					expectedCmd += "\n"
				}
				if string(tt.mockConn.WriteData) != expectedCmd {
					t.Errorf("Command written = %v, want %v", string(tt.mockConn.WriteData), expectedCmd)
				}
			}
		})
	}
}

func TestDaemonClient_IsRunning(t *testing.T) {
	tests := []struct {
		name          string
		mockConn      *MockConnection
		providerError error
		wantRunning   bool
		wantStatus    string
	}{
		{
			name: "daemon running",
			mockConn: &MockConnection{
				ReadData: "PONG\n",
			},
			wantRunning: true,
			wantStatus:  "running",
		},
		{
			name: "unexpected response",
			mockConn: &MockConnection{
				ReadData: "ERROR\n",
			},
			wantRunning: false,
			wantStatus:  "unexpected response: ERROR",
		},
		{
			name:          "connection error",
			providerError: errors.New("connection refused"),
			wantRunning:   false,
			wantStatus:    "not running: failed to connect to daemon: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &MockConnectionProvider{
				Conn:  tt.mockConn,
				Error: tt.providerError,
			}

			client := &DaemonClient{
				Provider: provider,
			}

			gotRunning, gotStatus := client.IsRunning(context.Background())
			if gotRunning != tt.wantRunning {
				t.Errorf("DaemonClient.IsRunning() running = %v, want %v", gotRunning, tt.wantRunning)
			}
			if gotStatus != tt.wantStatus {
				t.Errorf("DaemonClient.IsRunning() status = %v, want %v", gotStatus, tt.wantStatus)
			}
		})
	}
}
