package theme

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"strings"
)

// DisplayBanner prints a styled banner with the app name and description
func DisplayBanner(appCfg *config.AppConfig) {
	banner(fmt.Sprintf("Welcome to %s", appCfg.Name), 40, "Your AI assistant for the CLI")
}

// Banner prints a styled banner with the given title and optional subtitle
func banner(title string, width int, subtitle ...string) {
	theme := GetTheme()
	primary := theme.Primary()
	secondary := theme.Secondary()

	if width < len(title)+4 {
		width = len(title) + 4
	}

	// Check if subtitles need more width
	for _, sub := range subtitle {
		if len(sub)+4 > width {
			width = len(sub) + 4
		}
	}

	// Calculate padding for title
	titlePadding := (width - len(title) - 2) / 2
	titleLeftPadding := titlePadding
	titleRightPadding := titlePadding

	// Handle odd widths for title
	if (width-len(title)-2)%2 != 0 {
		titleRightPadding++
	}

	top := "╔" + strings.Repeat("═", width-2) + "╗"
	bottom := "╚" + strings.Repeat("═", width-2) + "╝"

	titleLeftSpaces := strings.Repeat(" ", titleLeftPadding)
	titleRightSpaces := strings.Repeat(" ", titleRightPadding)

	// Print top border
	primary.Println(top)

	// Print title
	primary.Println(fmt.Sprintf("║%s%s%s║", titleLeftSpaces, title, titleRightSpaces))

	// Print subtitles if provided
	if len(subtitle) > 0 {
		// Add a separator line if there are subtitles
		separator := "║" + strings.Repeat("─", width-2) + "║"
		primary.Println(separator)

		// Print each subtitle
		for _, sub := range subtitle {
			subPadding := (width - len(sub) - 2) / 2
			subLeftPadding := subPadding
			subRightPadding := subPadding

			// Handle odd widths for subtitle
			if (width-len(sub)-2)%2 != 0 {
				subRightPadding++
			}

			subLeftSpaces := strings.Repeat(" ", subLeftPadding)
			subRightSpaces := strings.Repeat(" ", subRightPadding)

			// Use secondary style for subtitle
			secondary.Println(fmt.Sprintf("║%s%s%s║", subLeftSpaces, sub, subRightSpaces))
		}
	}

	// Print bottom border
	primary.Println(bottom)
}
