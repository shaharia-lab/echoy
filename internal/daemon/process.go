package daemon

import "context"

type Process interface {
	Start() error
	Stop(ctx context.Context) error
}

// ManagedProcess defines the interface for sub-processes managed by the daemon.
type ManagedProcess interface {
	// Start initializes and starts the process.
	// It should return an error if initialization or startup fails.
	Start() error

	// Stop gracefully shuts down the process.
	// It should respect the provided context for deadlines or cancellations.
	Stop(ctx context.Context) error

	// Name returns a descriptive name for the process (useful for logging).
	Name() string
}
