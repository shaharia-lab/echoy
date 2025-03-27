package initializer

import (
	"github.com/shaharia-lab/echoy/internal/config"
	"gopkg.in/yaml.v3"
	"os"
)

// Implementation of ConfigManager interface

// LoadConfig loads the existing configuration
func (cm *DefaultConfigManager) LoadConfig() (config.Config, error) {
	var cfg config.Config

	configPath := config.GetConfigPath()
	configFile, err := os.ReadFile(configPath)
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
	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}

	// Marshal config to YAML
	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}

	// Write config file
	configPath := config.GetConfigPath()
	return os.WriteFile(configPath, yamlData, 0644)
}

// ConfigExists checks if a configuration file already exists
func (cm *DefaultConfigManager) ConfigExists() bool {
	return config.ConfigExists()
}
