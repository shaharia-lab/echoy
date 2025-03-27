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

type Filesystem struct {
	logger *logrus.Logger
	appCfg *config.AppConfig
	paths  map[PathType]string
}

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

	chatHistoryDBFilePath, err := s.CreateSQLiteDBFile(dataDir, "chat_history.db")
	if err != nil {
		return paths, err
	}
	paths[ChatHistoryDB] = chatHistoryDBFilePath

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
