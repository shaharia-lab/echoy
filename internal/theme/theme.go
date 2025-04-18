package theme

import (
	"github.com/fatih/color"
	"sync"
)

// Theme defines the interface for theming in the application
type Theme interface {
	// Primary returns the primary style
	Primary() StylePrinter

	// Secondary returns the secondary style
	Secondary() StylePrinter

	// Success returns the success style
	Success() StylePrinter

	// Error returns the error style
	Error() StylePrinter

	// Warning returns the warning style
	Warning() StylePrinter

	// Info returns the info style
	Info() StylePrinter

	// Subtle returns the subtle style
	Subtle() StylePrinter

	// Disabled returns the disabled style
	Disabled() StylePrinter

	// Custom returns a custom style by name
	Custom(name string) StylePrinter

	// IsEnabled reports if colors are enabled
	IsEnabled() bool

	// SetEnabled enables or disables color output
	SetEnabled(enabled bool)
}

// DefaultTheme represents the default theme implementation
type DefaultTheme struct {
	primary   *Style
	secondary *Style
	success   *Style
	error     *Style
	warning   *Style
	info      *Style
	subtle    *Style
	disabled  *Style
	custom    map[string]*Style
	enabled   bool
	mu        sync.RWMutex
}

// NewDefaultTheme creates a new default theme
func NewDefaultTheme() *DefaultTheme {
	return &DefaultTheme{
		primary:   NewStyle(color.FgHiCyan, 0, color.Bold),
		secondary: NewStyle(color.FgBlue, 0),
		success:   NewStyle(color.FgGreen, 0, color.Bold),
		error:     NewStyle(color.FgRed, 0, color.Bold),
		warning:   NewStyle(color.FgYellow, 0),
		info:      NewStyle(color.FgWhite, 0),
		subtle:    NewStyle(color.FgHiBlack, 0),
		disabled:  NewStyle(color.FgHiBlack, 0),
		custom:    make(map[string]*Style),
		enabled:   !color.NoColor,
	}
}

// NewProfessionalTheme creates a new professional theme
func NewProfessionalTheme() *DefaultTheme {
	return &DefaultTheme{
		// Primary: A calm blue that's visible but not overwhelming
		primary: NewStyle(color.FgBlue, 0, color.Bold),

		// Secondary: A subtle slate blue/gray
		secondary: NewStyle(color.FgHiBlue, 0),

		// Success: A muted green, less intense than the current bright green
		success: NewStyle(color.FgGreen, 0),

		// Error: A more muted red, still clear but less alarming
		error: NewStyle(color.FgRed, 0),

		// Warning: A softer amber color
		warning: NewStyle(color.FgYellow, 0),

		// Info: A clean white/light gray
		info: NewStyle(color.FgWhite, 0),

		// Subtle: A darker gray, still visible but clearly secondary
		subtle: NewStyle(color.FgHiBlack, 0),

		// Disabled: Same as subtle
		disabled: NewStyle(color.FgHiBlack, 0),

		// Initialize custom styles map
		custom: make(map[string]*Style),

		// Respect NO_COLOR environment variable
		enabled: !color.NoColor,
	}
}

// NewModernDarkTheme creates a new modern dark theme
func NewModernDarkTheme() *DefaultTheme {
	return &DefaultTheme{
		primary:   NewStyle(color.FgHiBlue, 0),
		secondary: NewStyle(color.FgBlue, 0),
		success:   NewStyle(color.FgHiGreen, 0),
		error:     NewStyle(color.FgHiRed, 0),
		warning:   NewStyle(color.FgHiYellow, 0),
		info:      NewStyle(color.FgHiWhite, 0),
		subtle:    NewStyle(color.FgWhite, 0),
		disabled:  NewStyle(color.FgHiBlack, 0),
		custom:    make(map[string]*Style),
		enabled:   !color.NoColor,
	}
}

func NewCorporateTheme() *DefaultTheme {
	return &DefaultTheme{
		// Primary: A professional navy blue
		primary: NewStyle(color.FgBlue, 0, color.Bold),

		// Secondary: A medium blue without bold
		secondary: NewStyle(color.FgBlue, 0),

		// Success: A business-appropriate green
		success: NewStyle(color.FgGreen, 0),

		// Error: A clear but not alarming red
		error: NewStyle(color.FgRed, 0),

		// Warning: A muted yellow/gold
		warning: NewStyle(color.FgYellow, 0),

		// Info: Clean white text
		info: NewStyle(color.FgWhite, 0),

		// Subtle: Gray for less important information
		subtle: NewStyle(color.FgHiBlack, 0),

		// Disabled: Light gray
		disabled: NewStyle(color.FgHiBlack, 0),

		// Initialize custom styles map
		custom: make(map[string]*Style),

		// Respect NO_COLOR environment variable
		enabled: !color.NoColor,
	}
}

// Primary returns the primary style
func (t *DefaultTheme) Primary() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.primary
}

// Secondary returns the secondary style
func (t *DefaultTheme) Secondary() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.secondary
}

// Success returns the success style
func (t *DefaultTheme) Success() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.success
}

// Error returns the error style
func (t *DefaultTheme) Error() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.error
}

// Warning returns the warning style
func (t *DefaultTheme) Warning() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.warning
}

// Info returns the info style
func (t *DefaultTheme) Info() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.info
}

// Subtle returns the subtle style
func (t *DefaultTheme) Subtle() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.subtle
}

// Disabled returns the disabled style
func (t *DefaultTheme) Disabled() StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.disabled
}

// Custom returns a custom style by name
func (t *DefaultTheme) Custom(name string) StylePrinter {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if style, ok := t.custom[name]; ok {
		return style
	}
	return t.info // Fallback to info style if the custom style doesn't exist
}

// RegisterCustomStyle registers a new custom style
func (t *DefaultTheme) RegisterCustomStyle(name string, style *Style) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.custom[name] = style
}

// IsEnabled reports if colors are enabled
func (t *DefaultTheme) IsEnabled() bool {
	return t.enabled && !color.NoColor
}

// SetEnabled enables or disables color output
func (t *DefaultTheme) SetEnabled(enabled bool) {
	t.enabled = enabled
}
