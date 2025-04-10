package logger

import (
	"context"
)

// NoopLogger is a logger implementation that does nothing.
// It's useful for testing or when logging should be disabled.
type NoopLogger struct{}

// NewNoopLogger creates a new no-op logger that discards all log messages.
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// WithField returns the same no-op logger.
func (l *NoopLogger) WithField(key string, value interface{}) Logger {
	return l
}

// WithFields returns the same no-op logger.
func (l *NoopLogger) WithFields(fields Fields) Logger {
	return l
}

// WithContext returns the same no-op logger.
func (l *NoopLogger) WithContext(ctx context.Context) Logger {
	return l
}

// Debug is a no-op.
func (l *NoopLogger) Debug(args ...interface{}) {}

// Debugf is a no-op.
func (l *NoopLogger) Debugf(format string, args ...interface{}) {}

// Info is a no-op.
func (l *NoopLogger) Info(args ...interface{}) {}

// Infof is a no-op.
func (l *NoopLogger) Infof(format string, args ...interface{}) {}

// Warn is a no-op.
func (l *NoopLogger) Warn(args ...interface{}) {}

// Warnf is a no-op.
func (l *NoopLogger) Warnf(format string, args ...interface{}) {}

// Error is a no-op.
func (l *NoopLogger) Error(args ...interface{}) {}

// Errorf is a no-op.
func (l *NoopLogger) Errorf(format string, args ...interface{}) {}

// Fatal is a no-op. In a real implementation, this would exit the program,
// but the no-op logger doesn't do that to avoid unexpected termination.
func (l *NoopLogger) Fatal(args ...interface{}) {}

// Fatalf is a no-op. In a real implementation, this would exit the program,
// but the no-op logger doesn't do that to avoid unexpected termination.
func (l *NoopLogger) Fatalf(format string, args ...interface{}) {}

// Flush is a no-op and returns nil because there's nothing to flush.
func (l *NoopLogger) Flush() error {
	return nil
}

// Ensure NoopLogger implements the Logger interface.
var _ Logger = (*NoopLogger)(nil)
