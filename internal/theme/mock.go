package theme

// MockTheme implements Theme interface for testing purposes
type MockTheme struct {
	// Default styles that will be returned by the theme methods
	primary   *Style
	secondary *Style
	success   *Style
	error     *Style
	warning   *Style
	info      *Style
	subtle    *Style
	disabled  *Style
	customs   map[string]*Style
	isEnabled bool
}

// NewMockTheme creates a new mock theme with default styles
func NewMockTheme() *MockTheme {
	return &MockTheme{
		primary:   NewStyle(0, 0), // Default empty style
		secondary: NewStyle(0, 0),
		success:   NewStyle(0, 0),
		error:     NewStyle(0, 0),
		warning:   NewStyle(0, 0),
		info:      NewStyle(0, 0),
		subtle:    NewStyle(0, 0),
		disabled:  NewStyle(0, 0),
		customs:   make(map[string]*Style),
		isEnabled: true,
	}
}

// Primary returns the primary style
func (m *MockTheme) Primary() *Style {
	return m.primary
}

// Secondary returns the secondary style
func (m *MockTheme) Secondary() *Style {
	return m.secondary
}

// Success returns the success style
func (m *MockTheme) Success() *Style {
	return m.success
}

// Error returns the error style
func (m *MockTheme) Error() *Style {
	return m.error
}

// Warning returns the warning style
func (m *MockTheme) Warning() *Style {
	return m.warning
}

// Info returns the info style
func (m *MockTheme) Info() *Style {
	return m.info
}

// Subtle returns the subtle style
func (m *MockTheme) Subtle() *Style {
	return m.subtle
}

// Disabled returns the disabled style
func (m *MockTheme) Disabled() *Style {
	return m.disabled
}

// Custom returns a custom style by name
func (m *MockTheme) Custom(name string) *Style {
	if style, ok := m.customs[name]; ok {
		return style
	}
	return NewStyle(0, 0) // Return default style if not found
}

// IsEnabled reports if colors are enabled
func (m *MockTheme) IsEnabled() bool {
	return m.isEnabled
}

// SetEnabled enables or disables color output
func (m *MockTheme) SetEnabled(enabled bool) {
	m.isEnabled = enabled
}

// SetPrimary sets the primary style for testing
func (m *MockTheme) SetPrimary(s *Style) {
	m.primary = s
}

// SetSecondary sets the secondary style for testing
func (m *MockTheme) SetSecondary(s *Style) {
	m.secondary = s
}

// SetSuccess sets the success style for testing
func (m *MockTheme) SetSuccess(s *Style) {
	m.success = s
}

// SetError sets the error style for testing
func (m *MockTheme) SetError(s *Style) {
	m.error = s
}

// SetWarning sets the warning style for testing
func (m *MockTheme) SetWarning(s *Style) {
	m.warning = s
}

// SetInfo sets the info style for testing
func (m *MockTheme) SetInfo(s *Style) {
	m.info = s
}

// SetSubtle sets the subtle style for testing
func (m *MockTheme) SetSubtle(s *Style) {
	m.subtle = s
}

// SetDisabled sets the disabled style for testing
func (m *MockTheme) SetDisabled(s *Style) {
	m.disabled = s
}

// SetCustom sets a custom style by name for testing
func (m *MockTheme) SetCustom(name string, s *Style) {
	m.customs[name] = s
}
