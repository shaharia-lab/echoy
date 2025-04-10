package logger

// LogLevel represents logging levels as strings
type LogLevel string

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in production
	DebugLevel LogLevel = "debug"

	// InfoLevel is the default logging priority
	InfoLevel LogLevel = "info"

	// WarnLevel logs are more important than Info, but don't need individual human review
	WarnLevel LogLevel = "warn"

	// ErrorLevel logs are high-priority
	ErrorLevel LogLevel = "error"

	// FatalLevel logs are particularly important errors, application will exit after logging
	FatalLevel LogLevel = "fatal"
)

// Logger defines the logging methods required by the application.
// Uses generic types to avoid coupling to a specific library implementation.
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Fatal(msg string, fields map[string]interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger

	Sync() error
}
