// Package llm contains the definition of the LLMProvider struct and the GetSupportedLLMProviders function.
package llm

import (
	"github.com/anthropics/anthropic-sdk-go"
)

// Provider represents a provider of large language models
type Provider struct {
	ID          string
	Name        string
	Description string
	ModelIDs    []string
}

// GetSupportedLLMProviders returns the list of supported LLM providers
func GetSupportedLLMProviders() []Provider {
	return []Provider{
		{
			ID:          "anthropic",
			Name:        "Anthropic",
			Description: "One of the leading AI/ML model providers",
			ModelIDs: []string{
				anthropic.ModelClaude3OpusLatest,
				anthropic.ModelClaude3_7SonnetLatest,
				anthropic.ModelClaude3_5HaikuLatest,
				anthropic.ModelClaude3_5SonnetLatest,
			},
		},
	}
}
