package init

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// ConfigureLLM configures LLM provider and model settings
func (i *Initializer) ConfigureLLM() {
	color.New(color.FgHiGreen, color.Bold).Println("\n🤖 Configure LLM Settings")

	// Define available providers list
	providers := []string{"openai", "anthropic", "local"}

	var selectedProvider string
	promptProvider := &survey.Select{
		Message: "Choose an LLM provider:",
		Options: providers,
		Default: i.Config.LLM.Provider,
	}
	survey.AskOne(promptProvider, &selectedProvider)
	i.Config.LLM.Provider = selectedProvider

	// Configure models based on the selected provider
	var modelOptions []string
	switch selectedProvider {
	case "openai":
		modelOptions = []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
	case "anthropic":
		modelOptions = []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}
	case "local":
		modelOptions = []string{"llama-3", "mistral"}
	}

	var selectedModel string
	promptModel := &survey.Select{
		Message: "Select a model:",
		Options: modelOptions,
		Default: i.Config.LLM.Model,
	}
	survey.AskOne(promptModel, &selectedModel)
	i.Config.LLM.Model = selectedModel

	// API token (with secure input)
	var apiToken string
	promptToken := &survey.Password{
		Message: "Enter your API token:",
		Help:    "This will be used to authenticate with the LLM provider",
	}
	if i.Config.LLM.Token != "" {
		color.Yellow("API token is already set. Press Enter to keep the existing token or enter a new one.")
	}
	survey.AskOne(promptToken, &apiToken)
	if apiToken != "" {
		i.Config.LLM.Token = apiToken
	}
}
