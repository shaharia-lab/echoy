package init

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// ConfigureAssistant configures assistant details
func (i *Initializer) ConfigureAssistant() {
	color.New(color.FgHiBlue, color.Bold).Println("📝 Assistant Details")

	var assistantName string
	promptAssistantName := &survey.Input{
		Message: "Provide a name of your assistant:",
		Help:    "Give your AI assistant a friendly name",
		Default: i.Config.Assistant.Name,
	}
	survey.AskOne(promptAssistantName, &assistantName)
	i.Config.Assistant.Name = assistantName
}
