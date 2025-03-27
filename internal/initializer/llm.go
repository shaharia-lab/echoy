package initializer

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/llm"
)

// ConfigureLLM configures LLM provider and model settings
func (i *Initializer) ConfigureLLM() error {
	i.cliTheme.GetCurrentTheme().Info().Println("\n🤖 Configure LLM Settings")

	// Get providers and find the current provider's name for default selection
	var providers = []string{}
	defaultProviderName := ""

	for _, provider := range llm.GetSupportedLLMProviders() {
		providers = append(providers, provider.Name)
		if provider.ID == i.Config.LLM.Provider {
			defaultProviderName = provider.Name
		}
	}

	// Collect provider information
	var selectedProvider string
	var providerID string

	promptProvider := &survey.Select{
		Message: "Choose an LLM provider:",
		Options: providers,
	}

	// Only set default if it's a valid option
	if defaultProviderName != "" {
		promptProvider.Default = defaultProviderName
	} else if len(providers) > 0 {
		// If no valid default, use first provider as default
		promptProvider.Default = providers[0]
	}

	err := survey.AskOne(promptProvider, &selectedProvider)
	if err != nil {
		return err
	}

	var modelOptions []string
	for _, provider := range llm.GetSupportedLLMProviders() {
		if provider.Name == selectedProvider {
			modelOptions = provider.ModelIDs
			providerID = provider.ID
			break
		}
	}

	if selectedProvider == "" {
		return fmt.Errorf("no valid LLM provider selected")
	}

	if len(modelOptions) == 0 {
		return fmt.Errorf("no models available for the selected provider")
	}

	// Collect model information
	var selectedModel string
	promptModel := &survey.Select{
		Message: "Select a model:",
		Options: modelOptions,
	}

	// Only set default model if it exists in the options
	modelExists := false
	for _, option := range modelOptions {
		if option == i.Config.LLM.Model {
			modelExists = true
			break
		}
	}

	if modelExists {
		promptModel.Default = i.Config.LLM.Model
	} else if len(modelOptions) > 0 {
		// If no valid default, use first model as default
		promptModel.Default = modelOptions[0]
	}

	err = survey.AskOne(promptModel, &selectedModel)
	if err != nil {
		return err
	}

	// Collect token information
	var apiToken string
	promptToken := &survey.Password{
		Message: "Enter your API token:",
		Help:    "This will be used to authenticate with the LLM provider",
	}

	if i.Config.LLM.Token != "" {
		color.Yellow("API token is already set. Press Enter to keep the existing token or enter a new one.")
	}

	err = survey.AskOne(promptToken, &apiToken)
	if err != nil {
		return err
	}

	// Only update config if all three settings are provided
	if apiToken == "" {
		if i.Config.LLM.Token == "" {
			return fmt.Errorf("API token is required")
		}
		// Using existing token
		apiToken = i.Config.LLM.Token
	}

	i.Config.LLM.Provider = providerID
	i.Config.LLM.Model = selectedModel
	i.Config.LLM.Token = apiToken

	var maxTokens int
	if i.Config.LLM.MaxTokens == 0 {
		i.Config.LLM.MaxTokens = 1000
	}

	promptMaxTokens := &survey.Input{
		Message: "Enter the maximum number of tokens:",
		Default: fmt.Sprintf("%d", i.Config.LLM.MaxTokens),
		Help:    "Defines the maximum length of the generated response.",
	}

	err = survey.AskOne(promptMaxTokens, &maxTokens)
	if err != nil {
		return err
	}

	var topP float64
	if i.Config.LLM.TopP == 0 {
		i.Config.LLM.TopP = 0.5
	}
	promptTopP := &survey.Input{
		Message: "Enter the top-p value (0.0 - 1.0):",
		Default: fmt.Sprintf("%f", i.Config.LLM.TopP),
		Help:    "Top-p sampling narrows the set of possible results to those whose probabilities sum up to this value.",
	}

	err = survey.AskOne(promptTopP, &topP)
	if err != nil {
		return err
	}

	var topK int
	if i.Config.LLM.TopK == 0 {
		i.Config.LLM.TopK = 50
	}
	promptTopK := &survey.Input{
		Message: "Enter the top-k value:",
		Default: fmt.Sprintf("%d", i.Config.LLM.TopK),
		Help:    "Limits the sampling to the top-k most likely results.",
	}

	err = survey.AskOne(promptTopK, &topK)
	if err != nil {
		return err
	}

	var temperature float64
	if i.Config.LLM.Temperature == 0 {
		i.Config.LLM.Temperature = 0.5
	}

	promptTemperature := &survey.Input{
		Message: "Enter the temperature value (0.0 - 1.0):",
		Default: fmt.Sprintf("%f", i.Config.LLM.Temperature),
		Help:    "Controls the creativity of the responses. Higher temperature values make results more varied.",
	}

	err = survey.AskOne(promptTemperature, &temperature)
	if err != nil {
		return err
	}

	i.Config.LLM.MaxTokens = maxTokens
	i.Config.LLM.TopP = topP
	i.Config.LLM.TopK = topK
	i.Config.LLM.Temperature = temperature

	return nil
}
