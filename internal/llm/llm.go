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
				anthropic.ModelClaude3_7SonnetLatest,
				anthropic.ModelClaude3_5HaikuLatest,
				anthropic.ModelClaude3_5SonnetLatest,
				anthropic.ModelClaude3OpusLatest,
			},
		},
		{
			ID:          "gemini",
			Name:        "Google Gemini",
			Description: "Google's Gemini model",
			ModelIDs: []string{
				"gemini-2.0-flash",
				"gemini-2.5-pro-exp-03-25",
				"gemini-2.0-flash-lite",
				"gemini-1.5-flash",
				"gemini-1.5-flash-8b",
				"gemini-1.5-pro",
			},
		},
	}
}
