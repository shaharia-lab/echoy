package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds the logger configuration options
type Config struct {
	LogLevel    Level
	LogFilePath string
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
	UseConsole  bool
	Development bool
}

// zapLogWriter implements io.Writer by sending all written data to the logger
type zapLogWriter struct {
	logger *ZapLogger
	level  Level
	prefix string
	buffer bytes.Buffer
}

// Write implements io.Writer interface
func (w *zapLogWriter) Write(p []byte) (n int, err error) {
	n, err = w.buffer.Write(p)
	if err != nil {
		return n, err
	}

	// Process each line separately
	for {
		line, err := w.buffer.ReadString('\n')
		if err == io.EOF {
			// Put the incomplete line back into the buffer
			w.buffer.WriteString(line)
			break
		}
		if err != nil {
			return n, err
		}

		// Remove trailing newline and any carriage returns
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		// Log the line with appropriate prefix
		logLine := w.prefix + line
		switch w.level {
		case DebugLevel:
			w.logger.Debug(logLine)
		case InfoLevel:
			w.logger.Info(logLine)
		case WarnLevel:
			w.logger.Warn(logLine)
		case ErrorLevel:
			w.logger.Error(logLine)
		case FatalLevel:
			// Avoid calling Fatal as it would terminate the program
			w.logger.Error("FATAL: " + logLine)
		default:
			w.logger.Info(logLine)
		}
	}

	return n, nil
}

// ZapLogger provides a concrete implementation of the Logger using zap.
type ZapLogger struct {
	zap *zap.Logger
	cfg Config
}

// Compile-time check to ensure ZapLogger implements the Logger interface.
var _ Logger = (*ZapLogger)(nil)

// NewZapLogger creates a new Zap logger satisfying the Logger interface.
func NewZapLogger(config Config) (Logger, error) {
	zapLogger, err := buildZapLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}

	return &ZapLogger{
		zap: zapLogger,
		cfg: config,
	}, nil
}

// buildZapLogger sets up the underlying zap logger instance.
func buildZapLogger(config Config) (*zap.Logger, error) {
	// Set defaults if not provided
	if config.MaxAgeDays <= 0 {
		config.MaxAgeDays = DefaultMaxAgeDays
	}
	if config.MaxSizeMB <= 0 {
		config.MaxSizeMB = DefaultMaxSizeMB
	}
	if config.MaxBackups < 0 {
		config.MaxBackups = DefaultMaxBackups
	}

	minLogLevel := getZapLogLevel(config.LogLevel)

	var encoderConfig zapcore.EncoderConfig
	if config.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
	}

	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	var cores []zapcore.Core

	newLumberjackWriter := func(filename string) (zapcore.WriteSyncer, error) {
		if filename == "" {
			return nil, fmt.Errorf("log filename is empty")
		}

		if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
			return nil, fmt.Errorf("failed to create log directory %s: %w", filepath.Dir(filename), err)
		}

		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   filename,
			MaxSize:    config.MaxSizeMB,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAgeDays,
			Compress:   true,
		}), nil
	}

	if config.LogFilePath != "" {
		writer, err := newLumberjackWriter(config.LogFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed creating info log writer: %w", err)
		}

		infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.DebugLevel && lvl <= zapcore.FatalLevel
		})
		cores = append(cores, zapcore.NewCore(jsonEncoder, writer, infoPriority))
	}

	if config.UseConsole {
		cores = append(cores, zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			minLogLevel,
		))
	}

	if len(cores) == 0 {
		fmt.Println("Warning: No logging outputs configured (files or console). Using no-op logger.")
		return zap.NewNop(), nil
	}

	core := zapcore.NewTee(cores...)

	var zapOpts []zap.Option
	zapOpts = append(zapOpts, zap.AddCaller())
	zapOpts = append(zapOpts, zap.AddCallerSkip(1))
	zapOpts = append(zapOpts, zap.AddStacktrace(zapcore.ErrorLevel))

	if config.Development {
		zapOpts = append(zapOpts, zap.Development())
	}

	return zap.New(core, zapOpts...), nil
}

// Helper to convert our Level type to Zap's level
func getZapLogLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// mapToZapFields converts our Fields type to zap.Field slice
func mapToZapFields(fields Fields) []zap.Field {
	if fields == nil {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

// WithField creates a new logger with a single field added
func (l *ZapLogger) WithField(key string, value interface{}) Logger {
	newZapLogger := l.zap.With(zap.Any(key, value))
	return &ZapLogger{
		zap: newZapLogger,
		cfg: l.cfg,
	}
}

// WithFields creates a new logger with additional fields
func (l *ZapLogger) WithFields(fields Fields) Logger {
	if len(fields) == 0 {
		return l
	}
	zapFields := mapToZapFields(fields)
	newZapLogger := l.zap.With(zapFields...)
	return &ZapLogger{
		zap: newZapLogger,
		cfg: l.cfg,
	}
}

// WithContext creates a new logger with the provided context
func (l *ZapLogger) WithContext(ctx context.Context) Logger {
	// You could extract trace IDs or other context values here if needed
	return l
}

// Debug logs a message at debug level
func (l *ZapLogger) Debug(args ...interface{}) {
	l.zap.Sugar().Debug(args...)
}

// Debugf logs a formatted message at debug level
func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.zap.Sugar().Debugf(format, args...)
}

// Info logs a message at info level
func (l *ZapLogger) Info(args ...interface{}) {
	l.zap.Sugar().Info(args...)
}

// Infof logs a formatted message at info level
func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

// Warn logs a message at warn level
func (l *ZapLogger) Warn(args ...interface{}) {
	l.zap.Sugar().Warn(args...)
}

// Warnf logs a formatted message at warn level
func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

// Error logs a message at error level
func (l *ZapLogger) Error(args ...interface{}) {
	l.zap.Sugar().Error(args...)
}

// Errorf logs a formatted message at error level
func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

// Fatal logs a message at fatal level
func (l *ZapLogger) Fatal(args ...interface{}) {
	l.zap.Sugar().Fatal(args...)
}

// Fatalf logs a formatted message at fatal level
func (l *ZapLogger) Fatalf(format string, args ...interface{}) {
	l.zap.Sugar().Fatalf(format, args...)
}

// Flush flushes any buffered log entries
func (l *ZapLogger) Flush() error {
	return l.zap.Sync()
}

// StdoutWriter returns an io.Writer that sends all written data to the logger at INFO level
func (l *ZapLogger) StdoutWriter() io.Writer {
	return &zapLogWriter{
		logger: l,
		level:  InfoLevel,
		prefix: "[stdout] ",
	}
}

// StderrWriter returns an io.Writer that sends all written data to the logger at ERROR level
func (l *ZapLogger) StderrWriter() io.Writer {
	return &zapLogWriter{
		logger: l,
		level:  ErrorLevel,
		prefix: "[stderr] ",
	}
}
