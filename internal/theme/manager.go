package theme

// Manager handles theme selection and management
type Manager struct {
	currentTheme Name
}

// NewManager creates a new theme manager with default settings
func NewManager() *Manager {
	return &Manager{
		currentTheme: Professional, // Using Professional as default as referenced in your code
	}
}

// SetThemeByName changes the current theme to the specified theme
func (m *Manager) SetThemeByName(name Name) {
	m.currentTheme = name

	// Call the existing global theme setter function to maintain compatibility
	// with the rest of the codebase during transition
	SetThemeByName(name)
}

// GetCurrentTheme returns the currently active theme
func (m *Manager) GetCurrentTheme() Theme {
	return GetTheme()
}
