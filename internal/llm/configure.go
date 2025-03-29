package llm

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/theme"
)

// ConfigureLLM configures LLM provider and model settings
func ConfigureLLM(themeManager *theme.Manager, config config.Config) error {
	themeManager.GetCurrentTheme().Info().Println("\n🤖 Configure LLM Settings")

	// Get providers and find the current provider's name for default selection
	var providers []string
	defaultProviderName := ""

	for _, provider := range GetSupportedLLMProviders() {
		providers = append(providers, provider.Name)
		if provider.ID == config.LLM.Provider {
			defaultProviderName = provider.Name
		}
	}

	var selectedProvider string
	var providerID string

	promptProvider := &survey.Select{
		Message: "Choose an LLM provider:",
		Options: providers,
	}

	if defaultProviderName != "" {
		promptProvider.Default = defaultProviderName
	} else if len(providers) > 0 {
		promptProvider.Default = providers[0]
	}

	err := survey.AskOne(promptProvider, &selectedProvider)
	if err != nil {
		return err
	}

	var modelOptions []string
	for _, provider := range GetSupportedLLMProviders() {
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

	var selectedModel string
	promptModel := &survey.Select{
		Message: "Select a model:",
		Options: modelOptions,
	}

	modelExists := false
	for _, option := range modelOptions {
		if option == config.LLM.Model {
			modelExists = true
			break
		}
	}

	if modelExists {
		promptModel.Default = config.LLM.Model
	} else if len(modelOptions) > 0 {
		promptModel.Default = modelOptions[0]
	}

	err = survey.AskOne(promptModel, &selectedModel)
	if err != nil {
		return err
	}

	var apiToken string
	promptToken := &survey.Password{
		Message: "Enter your API token:",
		Help:    "This will be used to authenticate with the LLM provider",
	}

	if config.LLM.Token != "" {
		color.Yellow("API token is already set. Press Enter to keep the existing token or enter a new one.")
	}

	err = survey.AskOne(promptToken, &apiToken)
	if err != nil {
		return err
	}

	if apiToken == "" {
		if config.LLM.Token == "" {
			return fmt.Errorf("API token is required")
		}
		apiToken = config.LLM.Token
	}

	config.LLM.Provider = providerID
	config.LLM.Model = selectedModel
	config.LLM.Token = apiToken

	var maxTokens int64
	if config.LLM.MaxTokens == 0 {
		config.LLM.MaxTokens = 1000
	}

	promptMaxTokens := &survey.Input{
		Message: "Enter the maximum number of tokens:",
		Default: fmt.Sprintf("%d", config.LLM.MaxTokens),
		Help:    "Defines the maximum length of the generated response.",
	}

	err = survey.AskOne(promptMaxTokens, &maxTokens)
	if err != nil {
		return err
	}

	var topP float64
	if config.LLM.TopP == 0 {
		config.LLM.TopP = 0.5
	}
	promptTopP := &survey.Input{
		Message: "Enter the top-p value (0.0 - 1.0):",
		Default: fmt.Sprintf("%f", config.LLM.TopP),
		Help:    "Top-p sampling narrows the set of possible results to those whose probabilities sum up to this value.",
	}

	err = survey.AskOne(promptTopP, &topP)
	if err != nil {
		return err
	}

	var topK int
	if config.LLM.TopK == 0 {
		config.LLM.TopK = 50
	}
	promptTopK := &survey.Input{
		Message: "Enter the top-k value:",
		Default: fmt.Sprintf("%d", config.LLM.TopK),
		Help:    "Limits the sampling to the top-k most likely results.",
	}

	err = survey.AskOne(promptTopK, &topK)
	if err != nil {
		return err
	}

	var temperature float64
	if config.LLM.Temperature == 0 {
		config.LLM.Temperature = 0.5
	}

	promptTemperature := &survey.Input{
		Message: "Enter the temperature value (0.0 - 1.0):",
		Default: fmt.Sprintf("%f", config.LLM.Temperature),
		Help:    "Controls the creativity of the responses. Higher temperature values make results more varied.",
	}

	err = survey.AskOne(promptTemperature, &temperature)
	if err != nil {
		return err
	}

	streaming := false
	if config.LLM.Streaming {
		streaming = true
	}

	promptStreaming := &survey.Confirm{
		Message: "Enable streaming mode?",
		Default: streaming,
		Help:    "Enables streaming mode for long-running conversations.",
	}

	err = survey.AskOne(promptStreaming, &streaming)

	config.LLM.MaxTokens = maxTokens
	config.LLM.TopP = topP
	config.LLM.TopK = topK
	config.LLM.Temperature = temperature
	config.LLM.Streaming = streaming

	return nil
}
