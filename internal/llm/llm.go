// Package llm contains the definition of the LLMProvider struct and the GetSupportedLLMProviders function.
package llm

// LLMProvider represents a provider of large language models
type LLMProvider struct {
	ID          string
	Name        string
	Description string
	ModelIDs    []string
}

// GetSupportedLLMProviders returns the list of supported LLM providers
func GetSupportedLLMProviders() []LLMProvider {
	return []LLMProvider{
		{
			ID:          "anthropic",
			Name:        "Anthropic",
			Description: "One of the leading AI/ML model providers",
			ModelIDs:    []string{"claude-2"},
		},
	}
}
