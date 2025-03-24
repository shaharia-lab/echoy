// Package filesystem contains the implementation of the Filesystem struct.
package filesystem

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

type PathType string

const (
	configYamlFileName = "config.yaml"

	AppDirectory    PathType = "app"
	CacheDirectory  PathType = "cache"
	ConfigDirectory PathType = "config"
	ConfigFilePath  PathType = "config_file"
	LogsDirectory   PathType = "logs"
	LogsFilePath    PathType = "log_file"
	DataDirectory   PathType = "data"
	ChatHistoryDB   PathType = "chat_history_db"
)

// Filesystem is a struct that contains the methods to interact with local storage.
type Filesystem struct {
	logger *logrus.Logger
	appCfg *config.AppConfig
	paths  map[PathType]string
}

// NewAppFilesystem creates a new Filesystem instance.
func NewAppFilesystem(appCfg *config.AppConfig) *Filesystem {
	return &Filesystem{
		appCfg: appCfg,
	}
}

func (s *Filesystem) EnsureAllPaths() (map[PathType]string, error) {
	paths := map[PathType]string{}

	appDirectory, err := s.EnsureAppDirectory()
	if err != nil {
		return paths, err
	}
	paths[AppDirectory] = appDirectory

	// create cache directory under app directory
	cacheDir := filepath.Join(appDirectory, "cache")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[CacheDirectory] = cacheDir

	// create config directory under app directory
	configDir := filepath.Join(appDirectory, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[ConfigDirectory] = configDir

	// create logs directory under app directory
	logsDir := filepath.Join(appDirectory, "logs")
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[LogsDirectory] = logsDir

	// create data directory under app directory
	dataDir := filepath.Join(appDirectory, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return paths, err
		}
	}
	paths[DataDirectory] = dataDir

	// create chat history database file under data directory
	chatHistoryDBFilePath, err := s.CreateSQLiteDBFile(dataDir, "chat_history.db")
	if err != nil {
		return paths, err
	}
	paths[ChatHistoryDB] = chatHistoryDBFilePath

	// create empty config file under config directory
	configFilePath := filepath.Join(configDir, configYamlFileName)
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		if _, err := os.Create(configFilePath); err != nil {
			return paths, err
		}
	}
	paths[ConfigFilePath] = configFilePath

	// create one empty log file under logs directory
	logFilePath := filepath.Join(logsDir, fmt.Sprintf("%s.log", strings.ToLower(s.appCfg.Name)))
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		if _, err := os.Create(logFilePath); err != nil {
			return paths, err
		}
	}
	paths[LogsFilePath] = logFilePath

	return paths, nil
}

func (s *Filesystem) EnsureAppDirectory() (string, error) {
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

func (s *Filesystem) CreateSQLiteDBFile(dataDirectory, fileName string) (string, error) {
	dbFilePath := filepath.Join(dataDirectory, fileName)
	if _, err := os.Stat(dbFilePath); err == nil {
		return dbFilePath, nil
	}

	file, err := os.Create(dbFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Initialize SQLite database
	sqliteDB, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return "", err
	}
	defer sqliteDB.Close()

	// Test connection to ensure SQLite file is valid
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
