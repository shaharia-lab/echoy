package init

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/banner"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/logger"
)

// Initializer handles the interactive setup process
type Initializer struct {
	Config       config.Config
	IsUpdateMode bool
	// Dependencies can be injected here for testing
	configManager ConfigManager
	log           *logger.Logger
	appConfig     *config.AppConfig
}

// ConfigManager interface for loading/saving configuration
type ConfigManager interface {
	LoadConfig() (config.Config, error)
	SaveConfig(config.Config) error
	ConfigExists() bool
}

// DefaultConfigManager implements ConfigManager with real file operations
type DefaultConfigManager struct{}

// NewInitializer creates a new initializer with default dependencies
func NewInitializer(log *logger.Logger, appCfg *config.AppConfig) *Initializer {
	return &Initializer{
		log:           log,
		appConfig:     appCfg,
		configManager: &DefaultConfigManager{},
	}
}

// WithConfigManager sets a custom config manager (useful for testing)
func (i *Initializer) WithConfigManager(cm ConfigManager) *Initializer {
	i.configManager = cm
	return i
}

// Run starts the interactive configuration process
func (i *Initializer) Run() error {
	banner.CLIBanner(i.appConfig).Display()

	var err error
	i.IsUpdateMode = i.configManager.ConfigExists()
	if i.IsUpdateMode {
		// Load existing config
		i.Config, err = i.configManager.LoadConfig()
		if err != nil {
			return fmt.Errorf("error loading configuration: %v", err)
		}

		color.New(color.FgHiMagenta, color.Bold).Println("🔄 Configuration Update Mode")
		color.New(color.FgHiWhite).Println("Existing configuration detected. Press Enter to keep current values, or provide new ones.")
	} else {
		i.Config = config.Config{} // Initialize with default values
		color.New(color.FgHiMagenta, color.Bold).Println("🔧 Initial Configuration")
		color.New(color.FgHiWhite).Println("Please configure your assistant for the first time. You can always change the configuration later.")
	}

	fmt.Println()

	// Configure different sections
	i.ConfigureAssistant()
	i.ConfigureUser()
	i.ConfigureTools()
	i.ConfigureLLM()

	// Save the configuration
	if err := i.configManager.SaveConfig(i.Config); err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	// Display completion message
	if i.IsUpdateMode {
		color.New(color.FgGreen, color.Bold).Println("\n✅ Configuration updated successfully!")
	} else {
		color.New(color.FgGreen, color.Bold).Println("\n✅ Echoy configured successfully!")
	}

	return nil
}
