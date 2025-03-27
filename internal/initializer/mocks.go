package initializer

import (
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/stretchr/testify/mock"
)

// MockConfigManager implements the ConfigManager interface for testing
type MockConfigManager struct {
	mock.Mock
	ExistingConfig config.Config
	ConfigSaved    config.Config
	ShouldExist    bool
}

func (m *MockConfigManager) LoadConfig() (config.Config, error) {
	args := m.Called()
	return args.Get(0).(config.Config), args.Error(1)
}

func (m *MockConfigManager) SaveConfig(cfg config.Config) error {
	args := m.Called(cfg)
	m.ConfigSaved = cfg
	return args.Error(0)
}

func (m *MockConfigManager) ConfigExists() bool {
	args := m.Called()
	return args.Bool(0)
}
