package config

import (
	"github.com/fatih/color"
	"os"
	"path/filepath"
)

// AssistantConfig represents the assistant configuration
type AssistantConfig struct {
	Name string `yaml:"name"`
}

// UserConfig represents the user information
type UserConfig struct {
	Name  string `yaml:"name,omitempty"`
	Email string `yaml:"email,omitempty"`
}

// DockerConfig represents Docker tool configuration
type DockerConfig struct {
	Enabled bool `yaml:"enabled"`
}

// GitConfig represents Git tool configuration
type GitConfig struct {
	Enabled              bool     `yaml:"enabled"`
	WhitelistedRepoPaths []string `yaml:"whitelisted_repo_paths,omitempty"`
	BlockedOperations    []string `yaml:"blocked_operation,omitempty"`
}

// SimpleEnabledConfig represents a simple tool configuration with just an enabled flag
type SimpleEnabledConfig struct {
	Enabled bool `yaml:"enabled"`
}

// ToolsConfig represents all tool configurations
type ToolsConfig struct {
	Docker DockerConfig        `yaml:"docker"`
	Git    GitConfig           `yaml:"git"`
	Sed    SimpleEnabledConfig `yaml:"sed"`
	Grep   SimpleEnabledConfig `yaml:"grep"`
	Cat    SimpleEnabledConfig `yaml:"cat"`
	Bash   SimpleEnabledConfig `yaml:"bash"`
}

// LLMConfig represents the LLM configuration
type LLMConfig struct {
	Provider string `yaml:"provider"`
	Model    string `yaml:"model"`
	Token    string `yaml:"token"`
}

// FrontendConfig represents the frontend configuration
type FrontendConfig struct {
	Enabled bool `yaml:"enabled"`
}

// Config represents the main configuration
type Config struct {
	Assistant AssistantConfig `yaml:"Assistant"`
	User      UserConfig      `yaml:"user"`
	Tools     ToolsConfig     `yaml:"tools"`
	LLM       LLMConfig       `yaml:"llm"`
	Frontend  FrontendConfig  `yaml:"frontend"`
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	configDir, err := os.UserHomeDir()
	if err != nil {
		color.Red("Error getting home directory: %v", err)
		os.Exit(1)
	}
	return filepath.Join(configDir, ".echoy", "config.yaml")
}

// ConfigExists checks if the configuration file exists
func ConfigExists() bool {
	configPath := GetConfigPath()
	_, err := os.Stat(configPath)
	return !os.IsNotExist(err)
}

// EnsureConfigDir creates the configuration directory if it doesn't exist
func EnsureConfigDir() error {
	configPath := GetConfigPath()
	configDir := filepath.Dir(configPath)
	return os.MkdirAll(configDir, 0755)
}
