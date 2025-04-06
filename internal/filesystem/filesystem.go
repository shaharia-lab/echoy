// Package filesystem provides the filesystem related operations like creating directories, files, etc.
package filesystem

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

// PathType represents a type of path
type PathType string

const (
	configYamlFileName = "config.yaml"

	// AppDirectory represents the application directory
	AppDirectory PathType = "app"

	// CacheDirectory represents the cache directory
	CacheDirectory PathType = "cache"

	// CacheWebuiBuild represents the cache webui build directory. i.e: cache/webui_build
	CacheWebuiBuild PathType = "cache_webui_build"

	// ConfigDirectory represents the config directory
	ConfigDirectory PathType = "config"

	// ConfigFilePath represents the config file path
	ConfigFilePath PathType = "config_file"

	// LogsDirectory represents the logs directory
	LogsDirectory PathType = "logs"

	// LogsFilePath represents the logs file path
	LogsFilePath PathType = "log_file"

	// DataDirectory represents the data directory
	DataDirectory PathType = "data"

	// ChatHistoryDB represents the chat history database
	ChatHistoryDB PathType = "chat_history_db"
)

// Filesystem provides filesystem related operations
type Filesystem struct {
	logger *logrus.Logger
	appCfg *config.AppConfig
	paths  map[PathType]string
}

// NewAppFilesystem creates a new filesystem instance
func NewAppFilesystem(appCfg *config.AppConfig) *Filesystem {
	return &Filesystem{
		appCfg: appCfg,
	}
}

// EnsureAllPaths ensures all required paths exist
func (s *Filesystem) EnsureAllPaths() (map[PathType]string, error) {
	paths := map[PathType]string{}

	appDirectory, err := s.ensureAppDirectory()
	if err != nil {
		return paths, err
	}
	paths[AppDirectory] = appDirectory

	cacheDir := filepath.Join(appDirectory, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[CacheDirectory] = cacheDir

	configDir := filepath.Join(appDirectory, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[ConfigDirectory] = configDir

	logsDir := filepath.Join(appDirectory, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[LogsDirectory] = logsDir

	dataDir := filepath.Join(appDirectory, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[DataDirectory] = dataDir

	chatHistoryDBFilePath, err := s.createChatHistoryDBFile(dataDir, "chat_history.db")
	if err != nil {
		return paths, err
	}
	paths[ChatHistoryDB] = chatHistoryDBFilePath

	systemFilePath := filepath.Join(dataDir, "system.json")
	if _, err := os.Stat(systemFilePath); os.IsNotExist(err) {
		file, err := os.Create(systemFilePath)
		if err != nil {
			return paths, err
		}
		defer file.Close()

		uid := uuid.New().String()
		systemData := fmt.Sprintf(`{"uuid": "%s"}`, uid)
		if _, err := file.Write([]byte(systemData)); err != nil {
			return paths, err
		}
	}

	// webui frontend
	frontendDir := filepath.Join(cacheDir, "webui_build")
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		if err := os.MkdirAll(frontendDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[CacheWebuiBuild] = frontendDir

	configFilePath := filepath.Join(configDir, configYamlFileName)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if _, err := os.Create(configFilePath); err != nil {
			return paths, err
		}
	}
	paths[ConfigFilePath] = configFilePath

	logFilePath := filepath.Join(logsDir, fmt.Sprintf("%s.log", strings.ToLower(s.appCfg.Name)))
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		if _, err := os.Create(logFilePath); err != nil {
			return paths, err
		}
	}
	paths[LogsFilePath] = logFilePath

	return paths, nil
}

func (s *Filesystem) ensureAppDirectory() (string, error) {
	homeDir, err := s.getUserHomeDirectory()
	if err != nil {
		return "", err
	}

	appDir := filepath.Join(homeDir, fmt.Sprintf(".%s", strings.ToLower(s.appCfg.Name)))

	if _, err := os.Stat(appDir); os.IsNotExist(err) {
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return "", err
		}
	}

	return appDir, nil
}

func (s *Filesystem) createChatHistoryDBFile(dataDirectory, fileName string) (string, error) {
	dbFilePath := filepath.Join(dataDirectory, fileName)
	if _, err := os.Stat(dbFilePath); err == nil {
		return dbFilePath, nil
	}

	file, err := os.Create(dbFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sqliteDB, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return "", err
	}
	defer sqliteDB.Close()

	if err := sqliteDB.Ping(); err != nil {
		return "", err
	}

	return dbFilePath, nil
}

func (s *Filesystem) getUserHomeDirectory() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return homeDir, nil
}

// GetSystemConfig returns the content of system.json file as a SystemConfig struct
func (s *Filesystem) GetSystemConfig() (*config.SystemConfig, error) {
	paths, err := s.EnsureAllPaths()
	if err != nil {
		return nil, err
	}

	dataDir := paths[DataDirectory]
	systemFilePath := filepath.Join(dataDir, "system.json")

	data, err := os.ReadFile(systemFilePath)
	if err != nil {
		return nil, err
	}

	var systemConfig config.SystemConfig
	if err := json.Unmarshal(data, &systemConfig); err != nil {
		return nil, err
	}

	return &systemConfig, nil
}
