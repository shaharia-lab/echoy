package theme

import (
	"log"
	"sync"
)

var (
	defaultTheme Theme
	once         sync.Once
)

// GetTheme returns the current theme instance
func GetTheme() Theme {
	once.Do(func() {
		log.Printf("Inside once.Do - defaultTheme is %v at %p", defaultTheme == nil, defaultTheme)
		if defaultTheme == nil {
			defaultTheme = NewDefaultTheme()
			log.Printf("Created new default theme at %p", defaultTheme)
		}
	})
	log.Printf("GetTheme returning: %T at %p", defaultTheme, defaultTheme)
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

// ResetTheme resets the theme instance to nil
func resetTheme() {
	defaultTheme = nil
	once = sync.Once{}
}
