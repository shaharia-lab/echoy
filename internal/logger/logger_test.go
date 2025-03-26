package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("default config values are set", func(t *testing.T) {
		logger, err := NewLogger(Config{})
		require.NoError(t, err)
		assert.Equal(t, InfoLevel, logger.cfg.LogLevel)
		assert.Equal(t, DefaultMaxSizeMB, logger.cfg.MaxSizeMB)
	})

	t.Run("custom config values are respected", func(t *testing.T) {
		config := Config{
			LogLevel:    DebugLevel,
			MaxSizeMB:   10,
			UseConsole:  true,
			Development: true,
		}
		logger, err := NewLogger(config)
		require.NoError(t, err)
		assert.Equal(t, DebugLevel, logger.cfg.LogLevel)
		assert.Equal(t, 10, logger.cfg.MaxSizeMB)
		assert.True(t, logger.cfg.UseConsole)
		assert.True(t, logger.cfg.Development)
	})

	t.Run("creates log directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		logDir := filepath.Join(tmpDir, "logs")
		logFile := filepath.Join(logDir, "app.log")

		_, err := NewLogger(Config{
			FilePath: logFile,
		})
		require.NoError(t, err)

		_, err = os.Stat(logDir)
		assert.NoError(t, err)
	})
}

func TestLogger_WithField(t *testing.T) {
	// Based on the function signature I found in the codebase
	logger, err := NewLogger(Config{UseConsole: true})
	require.NoError(t, err)

	newLogger := logger.WithField("key", "value")
	assert.NotNil(t, newLogger)
	assert.NotEqual(t, logger, newLogger, "WithField should return a new logger instance")
}

func TestLogLevels(t *testing.T) {
	testCases := []struct {
		name     string
		logLevel LogLevel
	}{
		{"debug level", DebugLevel},
		{"info level", InfoLevel},
		{"warn level", WarnLevel},
		{"error level", ErrorLevel},
		{"fatal level", FatalLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				LogLevel:   tc.logLevel,
				UseConsole: false,
			}

			logger, err := NewLogger(config)
			require.NoError(t, err)
			assert.Equal(t, tc.logLevel, logger.cfg.LogLevel)
		})
	}
}

func TestBuildZapLogger(t *testing.T) {
	t.Run("file only logger", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "app.log")

		config := Config{
			LogLevel:   InfoLevel,
			FilePath:   logFile,
			UseConsole: false,
		}

		zapLogger, err := buildZapLogger(config)
		require.NoError(t, err)
		assert.NotNil(t, zapLogger)

		zapLogger.Info("test message")

		_, err = os.Stat(logFile)
		assert.NoError(t, err)
	})

	t.Run("console only logger", func(t *testing.T) {
		config := Config{
			LogLevel:   InfoLevel,
			UseConsole: true,
		}

		zapLogger, err := buildZapLogger(config)
		require.NoError(t, err)
		assert.NotNil(t, zapLogger)
	})

	t.Run("both file and console logger", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "app.log")

		config := Config{
			LogLevel:   InfoLevel,
			FilePath:   logFile,
			UseConsole: true,
		}

		zapLogger, err := buildZapLogger(config)
		require.NoError(t, err)
		assert.NotNil(t, zapLogger)

		zapLogger.Info("test message")

		_, err = os.Stat(logFile)
		assert.NoError(t, err)
	})
}

func TestInvalidDirectory(t *testing.T) {
	invalidPath := filepath.Join(string(filepath.Separator), "proc", "non-existent-dir", "app.log")
	config := Config{
		FilePath: invalidPath,
	}

	logger, err := NewLogger(config)
	assert.Error(t, err)
	assert.Nil(t, logger)
}
