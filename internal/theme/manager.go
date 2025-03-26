package theme

// Manager handles theme selection and management
type Manager struct {
	currentTheme Theme
}

// NewManager creates a new theme manager with default settings
func NewManager(t Theme) *Manager {
	return &Manager{
		currentTheme: t,
	}
}

// GetCurrentTheme returns the currently active theme
func (m *Manager) GetCurrentTheme() Theme {
	return m.currentTheme
}
