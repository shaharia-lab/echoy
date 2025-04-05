package config

import (
	"fmt"
)

// Repository represents a GitHub repository
type Repository struct {
	Owner string
	Repo  string
}

type SystemConfig struct {
	UUID string `json:"uuid"`
}

// AppConfig represents the configuration for the application
type AppConfig struct {
	Name         string
	Repository   Repository
	Version      Version
	SystemConfig *SystemConfig
}

// Version represents the version information for the application
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
