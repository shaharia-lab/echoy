package config

import (
	"fmt"
)

type Repository struct {
	Owner string
	Repo  string
}

type AppConfig struct {
	Name       string
	Repository Repository
	Version    Version
}

type Version struct {
	Version string
	Commit  string
	Date    string
}

// VersionText returns the version information as a string
func (v *Version) VersionText() string {
	return fmt.Sprintf("v%s : %s (%s)", v.Version, v.Commit, v.Date)
}

// Option is a function that configures an AppConfig
type Option func(*AppConfig)

// WithVersion returns an option to set the version
func WithVersion(v Version) Option {
	return func(c *AppConfig) {
		c.Version = v
	}
}

// WithName returns an option to set the name
func WithName(name string) Option {
	return func(c *AppConfig) {
		c.Name = name
	}
}

// NewDefaultConfig creates a new AppConfig with default values and applies the given options
func NewDefaultConfig(opts ...Option) *AppConfig {
	cfg := &AppConfig{
		Name: "Echoy",
		Repository: Repository{
			Owner: "shaharia-lab",
			Repo:  "echoy",
		},
		Version: Version{
			Version: "0.0.0",
			Commit:  "unknown",
			Date:    "unknown",
		},
	}

	// Apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
