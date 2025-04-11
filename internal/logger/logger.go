// Package logger provides a common interface for logging libraries
package logger

const (
	// DefaultMaxSizeMB defines the default maximum size of a log file in megabytes before rotation
	DefaultMaxSizeMB = 100

	// DefaultMaxBackups defines the default maximum number of backup log files to keep
	DefaultMaxBackups = 3

	// DefaultMaxAgeDays defines the default maximum number of days to keep old log files
	DefaultMaxAgeDays = 30
)
