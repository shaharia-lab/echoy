package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "filesystem_test_*")
	require.NoError(t, err, "Failed to create temp directory")

	origHome := os.Getenv("HOME")

	err = os.Setenv("HOME", tempDir)
	require.NoError(t, err, "Failed to set HOME environment variable")

	cleanup := func() {
		os.Setenv("HOME", origHome)
		os.RemoveAll(tempDir)
	}
	return tempDir, cleanup
}

func TestNewAppFilesystem(t *testing.T) {
	appCfg := &config.AppConfig{
		Name: "TestApp",
	}

	fs := NewAppFilesystem(appCfg)
	assert.NotNil(t, fs, "Filesystem should not be nil")
}

func TestEnsureAppDirectory(t *testing.T) {
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	appCfg := &config.AppConfig{
		Name: "TestApp",
	}

	fs := NewAppFilesystem(appCfg)

	appDir, err := fs.ensureAppDirectory()
	assert.NoError(t, err, "ensureAppDirectory should not return an error")
	assert.NotEmpty(t, appDir, "App directory should not be empty")

	expectedPath := filepath.Join(tempHome, ".testapp")
	assert.Equal(t, expectedPath, appDir, "App directory path should match expected path")

	info, err := os.Stat(appDir)
	assert.NoError(t, err, "Should be able to stat the app directory")
	assert.True(t, info.IsDir(), "App path should be a directory")

	appDirAgain, err := fs.ensureAppDirectory()
	assert.NoError(t, err, "Second call to ensureAppDirectory should not return an error")
	assert.Equal(t, appDir, appDirAgain, "App directory path should be the same on second call")
}

func TestEnsureAllPaths(t *testing.T) {
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	appCfg := &config.AppConfig{
		Name: "TestApp",
	}

	fs := NewAppFilesystem(appCfg)

	paths, err := fs.EnsureAllPaths()
	assert.NoError(t, err, "EnsureAllPaths should not return an error")
	assert.NotNil(t, paths, "Paths map should not be nil")

	expectedAppDir := filepath.Join(tempHome, ".testapp")

	pathTests := []struct {
		pathType PathType
		subPath  string
		isDir    bool
	}{
		{AppDirectory, "", true},
		{CacheDirectory, "cache", true},
		{ConfigDirectory, "config", true},
		{LogsDirectory, "logs", true},
		{DataDirectory, "data", true},
		{ConfigFilePath, filepath.Join("config", "config.yaml"), false},
		{LogsFilePath, filepath.Join("logs", "testapp.log"), false},
		{ChatHistoryDB, filepath.Join("data", "chat_history.db"), false},
	}

	for _, tt := range pathTests {
		t.Run(string(tt.pathType), func(t *testing.T) {
			path, exists := paths[tt.pathType]
			assert.True(t, exists, "Path type %s should exist in paths map", tt.pathType)

			var expectedPath string
			if tt.subPath == "" {
				expectedPath = expectedAppDir
			} else {
				expectedPath = filepath.Join(expectedAppDir, tt.subPath)
			}
			assert.Equal(t, expectedPath, path, "Path for %s should match expected", tt.pathType)

			info, err := os.Stat(path)
			assert.NoError(t, err, "Should be able to stat %s", tt.pathType)

			if tt.isDir {
				assert.True(t, info.IsDir(), "Path %s should be a directory", tt.pathType)
			} else {
				assert.False(t, info.IsDir(), "Path %s should be a file", tt.pathType)
			}
		})
	}

	pathsAgain, err := fs.EnsureAllPaths()
	assert.NoError(t, err, "Second call to EnsureAllPaths should not return an error")
	assert.Equal(t, paths, pathsAgain, "Paths map should be the same on second call")
}

func TestCreateSQLiteDBFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sqlite_test_*")
	require.NoError(t, err, "Failed to create temp directory")
	defer os.RemoveAll(tempDir)

	appCfg := &config.AppConfig{
		Name: "TestApp",
	}

	fs := NewAppFilesystem(appCfg)

	dbFileName := "test.db"
	dbPath, err := fs.createChatHistoryDBFile(tempDir, dbFileName)
	assert.NoError(t, err, "createChatHistoryDBFile should not return an error")
	assert.NotEmpty(t, dbPath, "DB path should not be empty")

	expectedPath := filepath.Join(tempDir, dbFileName)
	assert.Equal(t, expectedPath, dbPath, "DB path should match expected path")

	info, err := os.Stat(dbPath)
	assert.NoError(t, err, "Should be able to stat the DB file")
	assert.False(t, info.IsDir(), "DB path should be a file")

	dbPathAgain, err := fs.createChatHistoryDBFile(tempDir, dbFileName)
	assert.NoError(t, err, "Second call to createChatHistoryDBFile should not return an error")
	assert.Equal(t, dbPath, dbPathAgain, "DB path should be the same on second call")
}

func TestErrorConditions(t *testing.T) {
	tempHome, cleanup := setupTestEnv(t)
	defer cleanup()

	readOnlyDir := filepath.Join(tempHome, "readonly")
	require.NoError(t, os.MkdirAll(readOnlyDir, 0755))

	require.NoError(t, os.Chmod(readOnlyDir, 0555))
	defer os.Chmod(readOnlyDir, 0755)

	appCfg := &config.AppConfig{
		Name: "TestApp",
	}

	fs := NewAppFilesystem(appCfg)

	_, err := fs.createChatHistoryDBFile(readOnlyDir, "test.db")
	assert.Error(t, err, "Creating DB file in read-only directory should fail")
}
