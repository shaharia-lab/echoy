package theme

import (
	"sync"
)

var (
	defaultTheme Theme
	once         sync.Once
)

// GetTheme returns the current theme instance
func GetTheme() Theme {
	once.Do(func() {
		// If no theme has been set yet, create the default one
		if defaultTheme == nil {
			defaultTheme = NewDefaultTheme()
		}
	})
	return defaultTheme
}

// SetTheme sets the theme instance for the application
func SetTheme(theme Theme) {
	defaultTheme = theme
}

func SetThemeByName(name Name) {
	var theme Theme

	switch name {
	case Default:
		theme = NewDefaultTheme()
	case Professional:
		theme = NewProfessionalTheme()
	case ModernDark:
		theme = NewModernDarkTheme()
	case Corporate:
		theme = NewCorporateTheme()
	default:
		theme = NewProfessionalTheme()
	}

	SetTheme(theme)
}
