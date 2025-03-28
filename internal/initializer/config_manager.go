package initializer

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"gopkg.in/yaml.v3"
	"os"
)

// LoadConfig loads the existing configuration
// LoadConfig loads the existing configuration or creates and loads default config if not found
func (cm *DefaultConfigManager) LoadConfig() (config.Config, error) {
	// Initialize with default configuration values
	c := config.Config{}
	defaultConfig := c.Default()

	if cm.configFilePath == "" {
		return defaultConfig, fmt.Errorf("config file path not set")
	}

	// If config file doesn't exist, save the default config
	if !cm.configFileExists() {
		err := cm.SaveConfig(defaultConfig)
		if err != nil {
			return config.Config{}, fmt.Errorf("failed to save default config: %w", err)
		}
		return defaultConfig, nil
	}

	// Config file exists, read it
	configFile, err := os.ReadFile(cm.configFilePath)
	if err != nil {
		return defaultConfig, fmt.Errorf("failed to read config file: %w", err)
	}

	// Check if the file is empty
	if len(configFile) == 0 {
		err := cm.SaveConfig(defaultConfig)
		if err != nil {
			return config.Config{}, fmt.Errorf("failed to save default config to empty file: %w", err)
		}
		return defaultConfig, nil
	}

	var cfg config.Config
	if err := yaml.Unmarshal(configFile, &cfg); err != nil {
		return defaultConfig, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// SaveConfig saves the configuration to disk
func (cm *DefaultConfigManager) SaveConfig(cfg config.Config) error {
	if cm.configFilePath == "" {
		return fmt.Errorf("config file path not set")
	}

	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(cm.configFilePath, yamlData, 0644)
}

// ConfigExists checks if a configuration file already exists
func (cm *DefaultConfigManager) configFileExists() bool {
	_, err := os.Stat(cm.configFilePath)
	return !os.IsNotExist(err)
}
