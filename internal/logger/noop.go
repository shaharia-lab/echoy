package logger

// NoOpLogger is a logger implementation that performs no actions (a "discard" logger).
// It implements the Logger interface but all logging methods are empty.
// Useful for testing environments or disabling logging entirely.
type NoOpLogger struct{}

// Discard is a ready-to-use NoOpLogger instance, suitable for use when logging output is not needed.
var Discard Logger = NoOpLogger{}

// NewNoOpLogger returns a logger instance that performs no operations.
// It's generally recommended to use the global 'Discard' variable.
func NewNoOpLogger() Logger {
	return Discard // Return the singleton instance
}

// Debug performs no action.
func (l NoOpLogger) Debug(msg string, fields map[string]interface{}) {}

// Info performs no action.
func (l NoOpLogger) Info(msg string, fields map[string]interface{}) {}

// Warn performs no action.
func (l NoOpLogger) Warn(msg string, fields map[string]interface{}) {}

// Error performs no action.
func (l NoOpLogger) Error(msg string, fields map[string]interface{}) {}

// Fatal performs no action.
// IMPORTANT: Unlike typical Fatal loggers, this No-Op version does NOT exit the application.
func (l NoOpLogger) Fatal(msg string, fields map[string]interface{}) {
}

// Debugf performs no action.
func (l NoOpLogger) Debugf(format string, args ...interface{}) {}

// Infof performs no action.
func (l NoOpLogger) Infof(format string, args ...interface{}) {}

// Warnf performs no action.
func (l NoOpLogger) Warnf(format string, args ...interface{}) {}

// Errorf performs no action.
func (l NoOpLogger) Errorf(format string, args ...interface{}) {}

// Fatalf performs no action.
func (l NoOpLogger) Fatalf(format string, args ...interface{}) {
}

// WithField returns the same NoOpLogger instance, allowing method chaining without effect.
func (l NoOpLogger) WithField(key string, value interface{}) Logger {
	return l
}

// WithFields returns the same NoOpLogger instance, allowing method chaining without effect.
func (l NoOpLogger) WithFields(fields map[string]interface{}) Logger {
	return l
}

// Sync performs no action and returns nil as there is nothing to sync.
func (l NoOpLogger) Sync() error {
	return nil
}

// This line doesn't execute code but will cause a compile error if
// NoOpLogger stops satisfying the Logger interface.
var _ Logger = NoOpLogger{}
