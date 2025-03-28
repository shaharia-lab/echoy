package cli

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/initializer"
	"github.com/shaharia-lab/echoy/internal/llm"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/goai"
)

// Container holds all application dependencies
type Container struct {
	Config             *config.AppConfig
	Filesystem         *filesystem.Filesystem
	Paths              map[filesystem.PathType]string
	Logger             *logger.Logger
	ThemeMgr           *theme.Manager
	ConfigManager      initializer.ConfigManager
	LLMService         llm.Service
	ChatHistoryService goai.ChatHistoryStorage
	ChatSession        *chat.Session
	ChatService        chat.Service
	Initializer        *initializer.Initializer
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

	if opts.Version == "" {
		return nil, fmt.Errorf("version is required")
	}

	if opts.Commit == "" {
		return nil, fmt.Errorf("commit is required")
	}

	if opts.Date == "" {
		return nil, fmt.Errorf("date is required")
	}

	if opts.LogLevel == "" {
		return nil, fmt.Errorf("log level is required")
	}

	if opts.Theme == nil {
		return nil, fmt.Errorf("theme is required")
	}

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

	container.ThemeMgr = theme.NewManager(opts.Theme, container.Config, &theme.StdoutWriter{})

	container.Filesystem = filesystem.NewAppFilesystem(container.Config)

	container.Paths, err = container.Filesystem.EnsureAllPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure all application paths: %w", err)
	}

	if container.Paths[filesystem.ConfigFilePath] == "" {
		return nil, fmt.Errorf("config file path is required")
	}

	loggerConfig := logger.Config{
		FilePath: container.Paths[filesystem.LogsFilePath],
		LogLevel: opts.LogLevel,
	}

	container.Logger, err = logger.NewLogger(loggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	container.Logger.Info("Logger initialized successfully")

	container.ConfigManager = initializer.NewDefaultConfigManager(container.Paths[filesystem.ConfigFilePath])
	cfg, err := container.ConfigManager.LoadConfig()
	if err != nil {
		container.Logger.Errorf(fmt.Sprintf("error loading configuration: %v", err))
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	container.Initializer = initializer.NewInitializer(container.Logger, container.Config, container.ThemeMgr, container.ConfigManager)

	container.LLMService, err = llm.NewLLMService(cfg.LLM)
	if err != nil {
		container.Logger.Errorf(fmt.Sprintf("error initializing LLM service: %v", err))
		return nil, fmt.Errorf("error initializing LLM service: %w", err)
	}

	container.ChatHistoryService = goai.NewInMemoryChatHistoryStorage()
	container.ChatService = chat.NewChatService(container.LLMService, container.ChatHistoryService)
	container.ChatSession, err = chat.NewChatSession(&cfg, container.ThemeMgr.GetCurrentTheme(), container.ChatService, container.ChatHistoryService)
	if err != nil {
		container.Logger.Errorf(fmt.Sprintf("error creating chat session: %v", err))
		return nil, fmt.Errorf("error creating chat session: %w", err)
	}

	return container, nil
}
