package telemetry

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/theme"
	"log"
)

// Configure sets up the telemetry configuration
func Configure(themeManager *theme.Manager, config *config.Config) error {
	themeManager.GetCurrentTheme().Info().Println("\n🤖 Enable anonymous telemetry event collection")
	themeManager.GetCurrentTheme().Info().Println("This helps us improve the product, prioritize features, and fix bugs.")
	themeManager.GetCurrentTheme().Info().Println("No personal data is collected, and you can opt out at any time.")
	themeManager.GetCurrentTheme().Info().Println("Help the project grow, and help the community.")

	enabled := false
	if config.UsageTracking.Enabled {
		enabled = true
	}

	promptTelemetryEnabling := &survey.Confirm{
		Message: "Enable anonymized telemetry event?",
		Default: enabled,
		Help:    "Enable anonymized telemetry event collection to help us improve the product.",
	}
	err := survey.AskOne(promptTelemetryEnabling, &enabled)
	if err != nil {
		log.Printf("Error while enabling/disabling telemetry: %v", err)
		return err
	}

	config.UsageTracking.Enabled = enabled
	return nil
}
