package cli

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/initializer"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
)

// Container holds all application dependencies
type Container struct {
	Config         *config.AppConfig
	Filesystem     *filesystem.Filesystem
	Paths          map[filesystem.PathType]string
	Logger         *logger.Logger
	ThemeMgr       *theme.Manager
	Initializer    *initializer.Initializer
	ConfigFromFile config.Config
}

// InitOptions contains options for initialization
type InitOptions struct {
	Version  string
	Commit   string
	Date     string
	LogLevel logger.LogLevel
	Theme    theme.Theme
}

// NewContainer creates and initializes all application dependencies
func NewContainer(opts InitOptions) (*Container, error) {
	container := &Container{}
	var err error

	defer func() {
		if container.Logger != nil {
			defer container.Logger.Sync()
		}
	}()

	container.Config = &config.AppConfig{
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

	container.ThemeMgr = theme.NewManager(theme.NewDefaultTheme(), container.Config, &theme.StdoutWriter{})

	container.Filesystem = filesystem.NewAppFilesystem(container.Config)

	container.Paths, err = container.Filesystem.EnsureAllPaths()
	if err != nil {
		return container, fmt.Errorf("failed to ensure all application paths: %w", err)
	}

	if container.Paths[filesystem.ConfigFilePath] == "" {
		return container, fmt.Errorf("config file path is required")
	}

	loggerConfig := logger.Config{
		FilePath: container.Paths[filesystem.LogsFilePath],
		LogLevel: logger.DebugLevel,
	}

	systemConfig, err := container.Filesystem.GetSystemConfig()
	if err != nil {
		return container, fmt.Errorf("failed to get system config: %w", err)
	}

	container.Config.SystemConfig = systemConfig

	container.Logger, err = logger.NewLogger(loggerConfig)
	if err != nil {
		return container, fmt.Errorf("failed to initialize logger: %w", err)
	}

	container.ConfigFromFile, err = initializer.NewDefaultConfigManager(container.Paths[filesystem.ConfigFilePath]).LoadConfig()
	if err != nil {
		container.Logger.Errorf(fmt.Sprintf("error loading configuration: %v", err))
		return container, fmt.Errorf("error loading configuration: %w", err)
	}

	configManager := initializer.NewDefaultConfigManager(container.Paths[filesystem.ConfigFilePath])
	container.Initializer = initializer.NewInitializer(container.Logger, container.Config, container.ThemeMgr, configManager)
	return container, nil
}
