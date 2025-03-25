package init

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/shaharia-lab/echoy/internal/cli"
)

// ConfigureAssistant configures assistant details
func (i *Initializer) ConfigureAssistant() error {
	i.cliTheme.Info().Println("📝 Assistant Details")

	var assistantName string
	promptAssistantName := &survey.Input{
		Message: "Name of your assistant:",
		Help:    "Give your AI assistant a friendly name",
		Default: i.Config.Assistant.Name,
	}
	err := survey.AskOne(promptAssistantName, &assistantName)
	if err != nil {
		return err
	}

	if assistantName == "" {
		assistantName = cli.GetAppConfig().Name
	}

	i.Config.Assistant.Name = assistantName
	return nil
}
