package cli

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/logger"
	"sync"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/theme"
)

var (
	// once ensures the initialization happens only once
	once sync.Once

	// initialized tracks whether Init has been called
	initialized bool

	// Internal app configuration
	appConfig *config.AppConfig

	// Application paths
	appPaths map[filesystem.PathType]string

	// Application logger
	appLogger *logger.Logger
)

// InitOptions contains options for initialization
type InitOptions struct {
	Version  string
	Commit   string
	Date     string
	LogLevel logger.LogLevel
}

// InitWithOptions initializes all CLI components with the given options
func InitWithOptions(opts InitOptions) error {
	var initErr error

	once.Do(func() {
		if err := initConfig(opts); err != nil {
			initErr = fmt.Errorf("failed to initialize config: %w", err)
			return
		}

		// Initialize filesystem
		if err := initFilesystem(); err != nil {
			initErr = fmt.Errorf("failed to initialize filesystem: %w", err)
			return
		}

		// Initialize logger
		if err := initLogger(opts); err != nil {
			initErr = fmt.Errorf("failed to initialize logger: %w", err)
			return
		}

		// Initialize theme
		initTheme()

		// Mark as initialized
		initialized = true
	})

	return initErr
}

// initConfig creates and configures the application config
func initConfig(opts InitOptions) error {
	// Create default config with version info
	appConfig = config.NewDefaultConfig(
		config.WithVersion(config.Version{
			Version: opts.Version,
			Commit:  opts.Commit,
			Date:    opts.Date,
		}),
	)

	return nil
}

// initFilesystem sets up the filesystem
func initFilesystem() error {
	var err error
	appPaths, err = filesystem.NewAppFilesystem(appConfig).EnsureAllPaths()
	if err != nil {
		return fmt.Errorf("failed to ensure all application paths: %w", err)
	}
	return nil
}

// initLogger sets up the application logger
func initLogger(opts InitOptions) error {
	var err error

	// Configure and create the logger
	loggerConfig := logger.Config{
		FilePath: appPaths[filesystem.LogsFilePath],
		LogLevel: opts.LogLevel,
	}

	appLogger, err = logger.NewLogger(loggerConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Log initialization success
	appLogger.Info("Logger initialized successfully")

	return nil
}

// initTheme sets up the theme system with professional defaults
func initTheme() {
	theme.SetTheme(theme.NewProfessionalTheme())
}
