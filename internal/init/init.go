package init

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
)

// Initializer handles the interactive setup process
type Initializer struct {
	Config       config.Config
	IsUpdateMode bool
	// Dependencies can be injected here for testing
	configManager ConfigManager
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
func NewInitializer() *Initializer {
	return &Initializer{
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
	PrintColorfulBanner()

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

// PrintColorfulBanner prints the application banner
func PrintColorfulBanner() {
	// Implementation of your banner printing function
	color.New(color.FgHiCyan, color.Bold).Println("╔════════════════════════════════════════╗")
	color.New(color.FgHiCyan, color.Bold).Println("║		  Welcome to Echoy			    ║")
	color.New(color.FgHiCyan, color.Bold).Println("╚════════════════════════════════════════╝")
}
