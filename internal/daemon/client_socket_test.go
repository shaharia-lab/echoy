package daemon

import (
	"context"
	"errors"
	daemonMocks "github.com/shaharia-lab/echoy/internal/daemon/mocks"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	// Copy data
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

func TestDaemonClient_Execute(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*daemonMocks.MockConnectionProvider, *MockConnection)
		cmd       string
		args      []string
		want      string
		wantErr   bool
	}{
		{
			name: "successful command without args",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				conn.ReadData = "SUCCESS\n"
				provider.EXPECT().Connect(mock.Anything).Return(conn, nil)
			},
			cmd:     "TEST",
			args:    nil,
			want:    "SUCCESS",
			wantErr: false,
		},
		{
			name: "successful command with args",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				conn.ReadData = "SUCCESS\n"
				provider.EXPECT().Connect(mock.Anything).Return(conn, nil)
			},
			cmd:     "TEST",
			args:    []string{"arg1", "arg2", "arg3"},
			want:    "SUCCESS",
			wantErr: false,
		},
		{
			name: "connection error",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				provider.EXPECT().Connect(mock.Anything).Return(nil, errors.New("connection refused"))
			},
			cmd:     "TEST",
			args:    nil,
			want:    "",
			wantErr: true,
		},
		{
			name: "write error",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				conn.WriteError = errors.New("write error")
				provider.EXPECT().Connect(mock.Anything).Return(conn, nil)
			},
			cmd:     "TEST",
			args:    nil,
			want:    "",
			wantErr: true,
		},
		{
			name: "read error",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				conn.ReadError = errors.New("read error")
				provider.EXPECT().Connect(mock.Anything).Return(conn, nil)
			},
			cmd:     "TEST",
			args:    nil,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := daemonMocks.NewMockConnectionProvider(t)
			mockConn := &MockConnection{}

			if tt.setupMock != nil {
				tt.setupMock(mockProvider, mockConn)
			}

			client := &Client{
				Provider: mockProvider,
			}

			got, err := client.Execute(context.Background(), tt.cmd, tt.args)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)

			if !tt.wantErr && mockConn.WriteData != nil {
				expectedCmd := tt.cmd
				if tt.args != nil && len(tt.args) > 0 {
					for _, arg := range tt.args {
						expectedCmd += " " + arg
					}
				}
				if !strings.HasSuffix(expectedCmd, "\n") {
					expectedCmd += "\n"
				}
				assert.Equal(t, expectedCmd, string(mockConn.WriteData))
			}
		})
	}
}

func TestDaemonClient_IsRunning(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*daemonMocks.MockConnectionProvider, *MockConnection)
		wantRunning bool
		wantStatus  string
	}{
		{
			name: "daemon running",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				conn.ReadData = "PONG\n"
				provider.EXPECT().Connect(mock.Anything).Return(conn, nil)
			},
			wantRunning: true,
			wantStatus:  "running",
		},
		{
			name: "unexpected response",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				conn.ReadData = "ERROR\n"
				provider.EXPECT().Connect(mock.Anything).Return(conn, nil)
			},
			wantRunning: false,
			wantStatus:  "unexpected response: ERROR",
		},
		{
			name: "connection error",
			setupMock: func(provider *daemonMocks.MockConnectionProvider, conn *MockConnection) {
				provider.EXPECT().Connect(mock.Anything).Return(nil, errors.New("connection refused"))
			},
			wantRunning: false,
			wantStatus:  "failed to connect to daemon: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := daemonMocks.NewMockConnectionProvider(t)
			mockConn := &MockConnection{}

			if tt.setupMock != nil {
				tt.setupMock(mockProvider, mockConn)
			}

			client := &Client{
				Provider: mockProvider,
			}

			gotRunning, gotStatus := client.IsRunning(context.Background())
			assert.Equal(t, tt.wantRunning, gotRunning)
			assert.Equal(t, tt.wantStatus, gotStatus)
		})
	}
}
