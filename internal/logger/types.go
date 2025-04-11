// Package logger provides a common interface for logging libraries
package logger

import (
	"context"
	"io"
)

// Level represents the severity level of a log message
type Level int

// Log levels
const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel

	ErrorKey = "error"
)

// Fields is a map of key-value pairs to add to a log entry
type Fields map[string]interface{}

// Logger is the interface that wraps the basic logging methods
type Logger interface {
	// WithField returns a new Logger with a single field added
	WithField(key string, value interface{}) Logger

	// WithFields returns a new Logger with additional fields
	WithFields(fields Fields) Logger

	// WithContext returns a new Logger with the given context
	WithContext(ctx context.Context) Logger

	// Debug logs a message at the debug level
	Debug(args ...interface{})

	// Debugf logs a formatted message at the debug level
	Debugf(format string, args ...interface{})

	// Info logs a message at the info level
	Info(args ...interface{})

	// Infof logs a formatted message at the info level
	Infof(format string, args ...interface{})

	// Warn logs a message at the warn level
	Warn(args ...interface{})

	// Warnf logs a formatted message at the warn level
	Warnf(format string, args ...interface{})

	// Error logs a message at the error level
	Error(args ...interface{})

	// Errorf logs a formatted message at the error level
	Errorf(format string, args ...interface{})

	// Fatal logs a message at the fatal level and exits
	Fatal(args ...interface{})

	// Fatalf logs a formatted message at the fatal level and exits
	Fatalf(format string, args ...interface{})

	// Flush ensures all pending log entries are written
	Flush() error

	// StdoutWriter returns an io.Writer that logs to stdout level
	StdoutWriter() io.Writer

	// StderrWriter returns an io.Writer that logs to error level
	StderrWriter() io.Writer
}
