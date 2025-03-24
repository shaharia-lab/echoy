package cli

import (
	"github.com/shaharia-lab/echoy/internal/theme"
	"sync"
)

var (
	// once ensures the initialization happens only once
	once sync.Once

	// initialized tracks whether Init has been called
	initialized bool
)

// Init initializes all CLI components (themes, loggers, etc.)
// This is safe to call multiple times - it will only execute once
func Init() {
	once.Do(func() {
		initTheme()
		initialized = true
	})
}

// initTheme sets up the theme system with professional defaults
func initTheme() {
	theme.SetTheme(theme.NewProfessionalTheme())
}

// IsInitialized returns whether the CLI has been initialized
func IsInitialized() bool {
	return initialized
}
