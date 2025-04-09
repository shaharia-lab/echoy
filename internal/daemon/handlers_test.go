package daemon

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestDefaultPingHandler(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		args        []string
		want        string
		expectError bool
	}{
		{
			name:        "Normal ping",
			ctx:         context.Background(),
			args:        []string{},
			want:        "PONG",
			expectError: false,
		},
		{
			name:        "Cancelled context",
			ctx:         cancelledContext(),
			args:        []string{},
			want:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultPingHandler(tt.ctx, tt.args)
			if (err != nil) != tt.expectError {
				t.Errorf("DefaultPingHandler() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.want {
				t.Errorf("DefaultPingHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeDefaultStatusHandler(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		configModifier func(*Config)
		setupCommands  map[string]CommandFunc
		expectContains []string
		expectError    bool
	}{
		{
			name: "Empty daemon status",
			ctx:  context.Background(),
			configModifier: func(cfg *Config) {
				cfg.MaxConnections = 100
			},
			setupCommands: nil,
			expectContains: []string{
				"Connections: 0 active",
				"Limit: 100",
				"Commands: 0 registered",
			},
			expectError: false,
		},
		{
			name: "Daemon with commands",
			ctx:  context.Background(),
			configModifier: func(cfg *Config) {
				cfg.MaxConnections = 200
			},
			setupCommands: map[string]CommandFunc{
				"TEST1": func(ctx context.Context, args []string) (string, error) { return "", nil },
				"TEST2": func(ctx context.Context, args []string) (string, error) { return "", nil },
			},
			expectContains: []string{
				"Connections: 0 active",
				"Limit: 200",
				"Commands: 2 registered",
				"TEST1, TEST2",
			},
			expectError: false,
		},
		{
			name:           "Cancelled context",
			ctx:            cancelledContext(),
			configModifier: func(cfg *Config) {},
			setupCommands:  nil,
			expectContains: []string{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
			}
			if tt.configModifier != nil {
				tt.configModifier(&cfg)
			}

			d, _ := createTestDaemon(t, cfg)

			if tt.setupCommands != nil {
				for name, cmd := range tt.setupCommands {
					d.RegisterCommand(name, cmd)
				}
			}

			handler := MakeDefaultStatusHandler(d)
			got, err := handler(tt.ctx, []string{})

			if (err != nil) != tt.expectError {
				t.Errorf("StatusHandler() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				for _, expected := range tt.expectContains {
					if !strings.Contains(got, expected) {
						t.Errorf("StatusHandler() output doesn't contain %q\nGot: %s", expected, got)
					}
				}
			}
		})
	}
}

func TestMakeDefaultStopHandler(t *testing.T) {
	t.Run("Stop command", func(t *testing.T) {
		d, _ := createTestDaemon(t, Config{
			Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		})

		stopCalled := make(chan struct{})

		originalStop := d.Stop

		var once sync.Once

		go func() {
			<-d.stopChan
			once.Do(func() {
				close(stopCalled)
			})
		}()

		handler := MakeDefaultStopHandler(d)

		result, err := handler(context.Background(), []string{})

		if err != nil {
			t.Errorf("StopHandler() error = %v", err)
		}

		if !strings.Contains(strings.ToLower(result), "stop initiated") {
			t.Errorf("StopHandler() = %v, should contain 'stop initiated'", result)
		}

		time.Sleep(50 * time.Millisecond)

		select {
		case <-stopCalled:
		case <-time.After(time.Second):
			t.Fatal("Daemon.Stop() did not close stopChan within timeout period")
		}

		originalStop()
	})
}

// Helper function for creating a cancelled context
func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
