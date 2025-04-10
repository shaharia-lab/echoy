package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds the logger configuration options
type Config struct {
	LogLevel LogLevel

	InfoFilePath  string
	WarnFilePath  string
	ErrorFilePath string

	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int

	UseConsole  bool
	Development bool
}

// ZapLogger provides a concrete implementation of the Logger using zap.
type ZapLogger struct {
	zap *zap.Logger
	cfg Config // Store config for reference if needed (e.g., for WithFields context)
}

// Compile-time check to ensure ZapLogger implements the Logger.
var _ Logger = (*ZapLogger)(nil)

// NewZapLogger creates a new Zap logger satisfying the Logger.
// It builds the underlying zap logger based on the configuration.
func NewZapLogger(config Config) (Logger, error) {
	zapLogger, err := buildZapLogger(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build zap logger: %w", err)
	}
	// Return the concrete type that implements the interface
	return &ZapLogger{
		zap: zapLogger,
		cfg: config,
	}, nil
}

// buildZapLogger sets up the underlying zap logger instance.
func buildZapLogger(config Config) (*zap.Logger, error) {
	if config.MaxAgeDays <= 0 {
		config.MaxAgeDays = DefaultMaxAgeDays
	}
	if config.MaxSizeMB <= 0 {
		config.MaxSizeMB = DefaultMaxSizeMB
	}
	if config.MaxBackups < 0 {
		config.MaxBackups = DefaultMaxBackups
	}
	if config.LogLevel == "" {
		config.LogLevel = DefaultLogLevel
	}

	minLogLevel := parseLogLevel(config.LogLevel)

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

	if config.ErrorFilePath != "" {
		writer, err := newLumberjackWriter(config.ErrorFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed creating error log writer: %w", err)
		}

		errorPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})

		cores = append(cores, zapcore.NewCore(jsonEncoder, writer, errorPriority))
	}

	if config.WarnFilePath != "" {
		writer, err := newLumberjackWriter(config.WarnFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed creating warn log writer: %w", err)
		}

		warnPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapcore.WarnLevel
		})
		cores = append(cores, zapcore.NewCore(jsonEncoder, writer, warnPriority))
	}

	if config.InfoFilePath != "" {
		writer, err := newLumberjackWriter(config.InfoFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed creating info log writer: %w", err)
		}

		infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == zapcore.InfoLevel
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

func mapToZapFields(fields map[string]interface{}) []zap.Field {
	if fields == nil {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return zapFields
}

// Debug logs a message at debug level
func (l *ZapLogger) Debug(msg string, fields map[string]interface{}) {
	l.zap.Debug(msg, mapToZapFields(fields)...)
}

// Info logs a message at info level
func (l *ZapLogger) Info(msg string, fields map[string]interface{}) {
	l.zap.Info(msg, mapToZapFields(fields)...)
}

// Warn logs a message at warn level
func (l *ZapLogger) Warn(msg string, fields map[string]interface{}) {
	l.zap.Warn(msg, mapToZapFields(fields)...)
}

// Error logs a message at error level
func (l *ZapLogger) Error(msg string, fields map[string]interface{}) {
	l.zap.Error(msg, mapToZapFields(fields)...)
}

// Fatal logs a message at fatal level
func (l *ZapLogger) Fatal(msg string, fields map[string]interface{}) {
	l.zap.Fatal(msg, mapToZapFields(fields)...)
}

// Debugf logs a formatted message at debug level
func (l *ZapLogger) Debugf(format string, args ...interface{}) {
	l.zap.Sugar().Debugf(format, args...)
}

// Infof logs a formatted message at info level
func (l *ZapLogger) Infof(format string, args ...interface{}) {
	l.zap.Sugar().Infof(format, args...)
}

// Warnf logs a formatted message at warn level
func (l *ZapLogger) Warnf(format string, args ...interface{}) {
	l.zap.Sugar().Warnf(format, args...)
}

// Errorf logs a formatted message at error level
func (l *ZapLogger) Errorf(format string, args ...interface{}) {
	l.zap.Sugar().Errorf(format, args...)
}

// Fatalf logs a formatted message at fatal level
func (l *ZapLogger) Fatalf(format string, args ...interface{}) {
	l.zap.Sugar().Fatalf(format, args...)
}

// WithField adds a single structured field to the logger context.
func (l *ZapLogger) WithField(key string, value interface{}) Logger {
	newZapLogger := l.zap.With(zap.Any(key, value))
	return &ZapLogger{
		zap: newZapLogger,
		cfg: l.cfg,
	}
}

// WithFields creates a new logger instance with additional fields
func (l *ZapLogger) WithFields(fields map[string]interface{}) Logger {
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

// Sync creates a new logger instance with a single field
func (l *ZapLogger) Sync() error {
	return l.zap.Sync()
}

func parseLogLevel(levelStr LogLevel) zapcore.Level {
	switch levelStr {
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
		fmt.Printf("Warning: Invalid log level '%s' specified, defaulting to 'info'\n", levelStr)
		return zapcore.InfoLevel
	}
}
