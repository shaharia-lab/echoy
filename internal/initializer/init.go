package initializer

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
)

// Initializer handles the interactive setup process
type Initializer struct {
	Config        config.Config
	IsUpdateMode  bool
	configManager ConfigManager
	log           *logger.Logger
	appConfig     *config.AppConfig
	cliTheme      *theme.Manager
}

// ConfigManager interface for loading/saving configuration
type ConfigManager interface {
	LoadConfig() (config.Config, error)
	SaveConfig(config.Config) error
	ConfigExists() bool
}

// DefaultConfigManager implements ConfigManager with real file operations
type DefaultConfigManager struct {
	configFilePath string
}

func NewDefaultConfigManager(configFilePath string) *DefaultConfigManager {
	return &DefaultConfigManager{configFilePath: configFilePath}
}

// NewInitializer creates a new initializer with default dependencies
func NewInitializer(log *logger.Logger, appCfg *config.AppConfig, theme *theme.Manager, configManager ConfigManager) *Initializer {
	return &Initializer{
		log:           log,
		appConfig:     appCfg,
		configManager: configManager,
		cliTheme:      theme,
	}
}

// WithConfigManager sets a custom config manager (useful for testing)
func (i *Initializer) WithConfigManager(cm ConfigManager) *Initializer {
	i.configManager = cm
	return i
}

// Run starts the interactive configuration process
func (i *Initializer) Run() error {
	i.log.Debug("Starting configuration process")

	var err error
	i.IsUpdateMode = i.configManager.ConfigExists()
	i.log.Debug(fmt.Sprintf("Update mode: %v", i.IsUpdateMode))

	if i.IsUpdateMode {
		i.Config, err = i.configManager.LoadConfig()
		if err != nil {
			i.log.Error(fmt.Sprintf("error loading configuration: %v", err))
			return fmt.Errorf("error loading configuration: %v", err)
		}

		i.cliTheme.GetCurrentTheme().Primary().Println("🔄 Configuration Update Mode")
		i.cliTheme.GetCurrentTheme().Warning().Println("You are about to update your existing configuration. Press Enter to keep current values, or provide new ones.")
	} else {
		i.Config = config.Config{}
		i.cliTheme.GetCurrentTheme().Primary().Println("🔧 Initial Configuration")
		i.cliTheme.GetCurrentTheme().Info().Println("Please configure your assistant for the first time. You can always change the configuration later.")
	}

	fmt.Println()

	err = i.ConfigureAssistant("Ehcoy")
	if err != nil {
		i.log.Error(fmt.Sprintf("error configuring assistant: %v", err))
		return fmt.Errorf("error configuring assistant: %v", err)
	}

	err = i.ConfigureUser()
	if err != nil {
		i.log.Error(fmt.Sprintf("error configuring user: %v", err))
		return fmt.Errorf("error configuring user: %v", err)
	}

	err = i.ConfigureLLM()
	if err != nil {
		i.log.Error(fmt.Sprintf("error configuring LLM: %v", err))
		return fmt.Errorf("error configuring LLM: %v", err)
	}

	i.log.Debug(fmt.Sprintf("Saving configuration: %v", i.Config))
	if err := i.configManager.SaveConfig(i.Config); err != nil {
		i.log.Error(fmt.Sprintf("error saving configuration: %v", err))
		return fmt.Errorf("error saving configuration: %v", err)
	}

	i.log.Debug("Configuration process complete")
	i.cliTheme.GetCurrentTheme().Success().Println("\n✅ Configuration updated successfully!")
	i.cliTheme.GetCurrentTheme().Info().Println("Run 'echoy chat' to start an interactive chat session.")
	return nil
}
