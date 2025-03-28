package initializer

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"gopkg.in/yaml.v3"
	"os"
)

// LoadConfig loads the existing configuration
func (cm *DefaultConfigManager) LoadConfig() (config.Config, error) {
	var cfg config.Config

	if cm.configFilePath == "" {
		return cfg, fmt.Errorf("config file path not set")
	}

	configFile, err := os.ReadFile(cm.configFilePath)
	if err != nil {
		return cfg, err
	}

	if err := yaml.Unmarshal(configFile, &cfg); err != nil {
		return cfg, err
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
func (cm *DefaultConfigManager) ConfigExists() bool {
	return config.ConfigExists()
}
