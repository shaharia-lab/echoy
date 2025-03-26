package init

import (
	"github.com/AlecAivazis/survey/v2"
)

// ConfigureAssistant configures assistant details
func (i *Initializer) ConfigureAssistant(assistantName string) error {
	i.cliTheme.GetCurrentTheme().Primary().Println("📝 Assistant Details")

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
		assistantName = assistantName
	}

	i.Config.Assistant.Name = assistantName
	return nil
}
