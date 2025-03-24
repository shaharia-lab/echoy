package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// MaxLogRetentionDays defines how long to keep logs (15 days)
	MaxLogRetentionDays = 15

	// DefaultMaxSizeMB defines the default maximum size of a log file in megabytes before rotation
	DefaultMaxSizeMB = 100
)

// LogLevel represents logging levels
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

// Config holds the logger configuration options
type Config struct {
	// LogLevel sets the minimum level of severity for logging messages
	LogLevel LogLevel
	// FilePath specifies where to write the log file
	FilePath string
	// MaxSizeMB is the maximum size in megabytes of the log file before it gets rotated
	MaxSizeMB int
	// UseConsole determines if logs are also written to console
	UseConsole bool
	// Development puts the logger in development mode, which changes the behavior of DPanicLevel
	Development bool
}

// Logger wraps the zap logger functionality
type Logger struct {
	zap *zap.Logger
	cfg Config
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config Config) (*Logger, error) {
	// Set default values if not provided
	if config.LogLevel == "" {
		config.LogLevel = InfoLevel
	}
	if config.MaxSizeMB == 0 {
		config.MaxSizeMB = DefaultMaxSizeMB
	}

	// Set up the logger
	logger, err := buildZapLogger(config)
	if err != nil {
		return nil, err
	}

	return &Logger{
		zap: logger,
		cfg: config,
	}, nil
}

// buildZapLogger sets up the zap logger with the provided configuration
func buildZapLogger(config Config) (*zap.Logger, error) {
	// Create directory for log file if it doesn't exist
	if config.FilePath != "" {
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Parse log level
	level := zapcore.InfoLevel
	switch config.LogLevel {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	}

	// Set encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create cores - we'll have at least one for file logging
	var cores []zapcore.Core

	// Add file logging if path is provided
	if config.FilePath != "" {
		// Set up lumberjack for log rotation
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSizeMB,
			MaxBackups: 0, // Keep all backups within MaxAge
			MaxAge:     MaxLogRetentionDays,
			Compress:   true,
		})

		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			level,
		))
	}

	// Add console logging if enabled
	if config.UseConsole {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		))
	}

	// Combine cores
	core := zapcore.NewTee(cores...)

	// Create the logger
	var zapOpts []zap.Option
	if config.Development {
		zapOpts = append(zapOpts, zap.Development())
	}
	zapOpts = append(zapOpts, zap.AddCaller())
	zapOpts = append(zapOpts, zap.AddCallerSkip(1))

	return zap.New(core, zapOpts...), nil
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.zap.With(zap.Any(key, value))
	return &Logger{
		zap: newLogger,
		cfg: l.cfg,
	}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	newLogger := l.zap.With(zapFields...)
	return &Logger{
		zap: newLogger,
		cfg: l.cfg,
	}
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}

// Debugf logs a formatted message at debug level
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zap.Sugar().Debugf(format, args...)
}

// Info logs a message at info level
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}

// Infof logs a formatted message at info level
func (l *Logger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}

// Warnf logs a formatted message at warn level
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}

// Errorf logs a formatted message at error level
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}

// Fatalf logs a formatted message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zap.Sugar().Fatalf(format, args...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zap.Sync()
}
