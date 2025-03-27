package theme

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"strings"
)

// Manager handles theme selection and management
type Manager struct {
	currentTheme Theme
	appConfig    *config.AppConfig
	writer       Writer
}

// NewManager creates a new theme manager with default settings
func NewManager(t Theme, appConfig *config.AppConfig, writer Writer) *Manager {
	if writer == nil {
		writer = &StdoutWriter{}
	}

	return &Manager{
		currentTheme: t,
		appConfig:    appConfig,
		writer:       writer,
	}
}

// WithWriter sets a custom writer for the manager
func (m *Manager) WithWriter(w Writer) *Manager {
	m.writer = w
	return m
}

// GetCurrentTheme returns the currently active theme
func (m *Manager) GetCurrentTheme() Theme {
	return m.currentTheme
}

// DisplayBanner prints a styled banner with the app name and description
func (m *Manager) DisplayBanner(title string, width int, subtitle ...string) {
	primary := m.currentTheme.Primary()
	secondary := m.currentTheme.Secondary()

	if width < len(title)+4 {
		width = len(title) + 4
	}

	for _, sub := range subtitle {
		if len(sub)+4 > width {
			width = len(sub) + 4
		}
	}

	titlePadding := (width - len(title) - 2) / 2
	titleLeftPadding := titlePadding
	titleRightPadding := titlePadding

	if (width-len(title)-2)%2 != 0 {
		titleRightPadding++
	}

	top := "╔" + strings.Repeat("═", width-2) + "╗"
	bottom := "╚" + strings.Repeat("═", width-2) + "╝"

	titleLeftSpaces := strings.Repeat(" ", titleLeftPadding)
	titleRightSpaces := strings.Repeat(" ", titleRightPadding)

	primary.Println(top)

	primary.Println(fmt.Sprintf("║%s%s%s║", titleLeftSpaces, title, titleRightSpaces))

	if len(subtitle) > 0 {
		separator := "║" + strings.Repeat("─", width-2) + "║"
		primary.Println(separator)

		for _, sub := range subtitle {
			subPadding := (width - len(sub) - 2) / 2
			subLeftPadding := subPadding
			subRightPadding := subPadding

			if (width-len(sub)-2)%2 != 0 {
				subRightPadding++
			}

			subLeftSpaces := strings.Repeat(" ", subLeftPadding)
			subRightSpaces := strings.Repeat(" ", subRightPadding)

			secondary.Println(fmt.Sprintf("║%s%s%s║", subLeftSpaces, sub, subRightSpaces))
		}
	}

	primary.Println(bottom)
}
