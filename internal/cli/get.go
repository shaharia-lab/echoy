package cli

import (
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/filesystem"
	"github.com/shaharia-lab/echoy/internal/logger"
)

// IsInitialized returns whether the CLI has been initialized
func IsInitialized() bool {
	return initialized
}

// GetAppConfig returns the application configuration
func GetAppConfig() *config.AppConfig {
	return appConfig
}

// GetAppPaths returns the application paths
func GetAppPaths() map[filesystem.PathType]string {
	return appPaths
}

// GetLogger returns the application logger
func GetLogger() *logger.Logger {
	return appLogger
}
