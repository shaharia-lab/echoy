package cli

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
)

// Container holds all application dependencies
type Container struct {
	Config     *config.AppConfig
	Filesystem *filesystem.Filesystem
	Paths      map[filesystem.PathType]string
	Logger     *logger.Logger
	ThemeMgr   *theme.Manager
}

// InitOptions contains options for initialization
type InitOptions struct {
	Version  string
	Commit   string
	Date     string
	LogLevel logger.LogLevel
	Theme    theme.Name
}

// NewContainer creates and initializes all application dependencies
func NewContainer(opts InitOptions) (*Container, error) {
	container := &Container{}
	var err error

	// Initialize config
	if err = initConfig(container, opts); err != nil {
		return nil, fmt.Errorf("failed to initialize config: %w", err)
	}

	// Initialize filesystem
	if err = initFilesystem(container); err != nil {
		return nil, fmt.Errorf("failed to initialize filesystem: %w", err)
	}

	// Initialize logger
	if err = initLogger(container, opts); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize theme
	initTheme(container, opts)

	return container, nil
}

// initConfig creates and configures the application config
func initConfig(c *Container, opts InitOptions) error {
	c.Config = &config.AppConfig{
		Name: "Echoy",
		Repository: config.Repository{
			Owner: "shaharia-lab",
			Repo:  "echoy",
		},
		Version: config.Version{
			Version: opts.Version,
			Commit:  opts.Commit,
			Date:    opts.Date,
		},
	}

	return nil
}

// initFilesystem sets up the filesystem
func initFilesystem(c *Container) error {
	c.Filesystem = filesystem.NewAppFilesystem(c.Config)

	var err error
	c.Paths, err = c.Filesystem.EnsureAllPaths()
	if err != nil {
		return fmt.Errorf("failed to ensure all application paths: %w", err)
	}

	return nil
}

// initLogger sets up the application logger
func initLogger(c *Container, opts InitOptions) error {
	// Configure and create the logger
	loggerConfig := logger.Config{
		FilePath: c.Paths[filesystem.LogsFilePath],
		LogLevel: opts.LogLevel,
	}

	var err error
	c.Logger, err = logger.NewLogger(loggerConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Log initialization success
	c.Logger.Info("Logger initialized successfully")

	return nil
}

// initTheme sets up the theme system with professional defaults
func initTheme(c *Container, opts InitOptions) {
	c.ThemeMgr = theme.NewManager()

	themeName := opts.Theme
	if themeName == "" {
		themeName = theme.Professional
	}

	c.ThemeMgr.SetThemeByName(themeName)
}
