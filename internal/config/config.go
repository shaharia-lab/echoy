package config

// AssistantConfig represents the assistant configuration
type AssistantConfig struct {
	Name string `yaml:"name"`
}

// UserConfig represents the user information
type UserConfig struct {
	Name string `yaml:"name,omitempty"`
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
	Provider    string  `yaml:"provider"`
	Model       string  `yaml:"model"`
	Token       string  `yaml:"token"`
	MaxTokens   int64   `yaml:"max_tokens"`
	Streaming   bool    `yaml:"streaming"`
	TopP        float64 `yaml:"top_p"`
	Temperature float64 `yaml:"temperature"`
	TopK        int     `yaml:"top_k"`
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

func (c *Config) Default() Config {
	return Config{
		Assistant: AssistantConfig{
			Name: "Echoy",
		},
		User: UserConfig{
			Name: "",
		},
		Tools: ToolsConfig{
			Docker: DockerConfig{
				Enabled: false,
			},
			Git: GitConfig{
				Enabled:              true,
				WhitelistedRepoPaths: []string{},
				BlockedOperations:    []string{},
			},
			Sed: SimpleEnabledConfig{
				Enabled: true,
			},
			Grep: SimpleEnabledConfig{
				Enabled: true,
			},
			Cat: SimpleEnabledConfig{
				Enabled: true,
			},
			Bash: SimpleEnabledConfig{
				Enabled: true,
			},
		},
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-3.5-turbo",
			Token:       "",
			MaxTokens:   4096,
			Streaming:   true,
			TopP:        1.0,
			Temperature: 0.7,
			TopK:        50,
		},
	}
}
