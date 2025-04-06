// Package llm contains the definition of the LLMProvider struct and the GetSupportedLLMProviders function.
package llm

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/openai/openai-go"
)

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
			Models: []Model{
				{
					Name:        "Claude 3.5 Haiku Latest",
					Description: "Fast and cost-effective model",
					ModelID:     anthropic.ModelClaude3_5HaikuLatest,
				},
				{
					Name:        "Claude 3.5 Haiku 2024-10-22",
					Description: "Fast and cost-effective model",
					ModelID:     anthropic.ModelClaude3_5Haiku20241022,
				},
				{
					Name:        "Claude 3.7 Sonnet",
					Description: "Most intelligent model from Anthropic",
					ModelID:     anthropic.ModelClaude3_7SonnetLatest,
				},
				{
					Name:        "Claude 3.5 Sonnet Latest",
					Description: "Our most intelligent model",
					ModelID:     anthropic.ModelClaude3_5SonnetLatest,
				},
				{
					Name:        "Claude 3.5 Sonnet 2024-10-22",
					Description: "Our most intelligent model",
					ModelID:     anthropic.ModelClaude3_5Sonnet20241022,
				},
				{
					Name:        "Claude 3.5 Sonnet 2024-06-20",
					Description: "Our previous most intelligent model",
					ModelID:     anthropic.ModelClaude_3_5_Sonnet_20240620,
				},
				{
					Name:        "Claude 3 Opus Latest",
					Description: "Excels at writing and complex tasks",
					ModelID:     anthropic.ModelClaude3OpusLatest,
				},
				{
					Name:        "Claude 3 Opus 2024-02-29",
					Description: "Excels at writing and complex tasks",
					ModelID:     anthropic.ModelClaude_3_Opus_20240229,
				},
				{
					Name:        "Claude 3 Sonnet 2024-02-29",
					Description: "Balance of speed and intelligence",
					ModelID:     anthropic.ModelClaude_3_Sonnet_20240229,
				},
				{
					Name:        "Claude 3 Haiku 2024-03-07",
					Description: "Our previous fast and cost-effective",
					ModelID:     anthropic.ModelClaude_3_Haiku_20240307,
				},
				{
					Name:        "Claude 2.1",
					Description: "Powerful language model for general-purpose tasks",
					ModelID:     anthropic.ModelClaude_2_1,
				},
				{
					Name:        "Claude 2.0",
					Description: "Advanced language model optimized for reliability and thoughtful responses",
					ModelID:     anthropic.ModelClaude_2_0,
				},
			},
		},
		{
			ID:          "gemini",
			Name:        "Google Gemini",
			Description: "Google's Gemini model",
			Models: []Model{
				{
					Name:        "Gemini 2.0 Flash",
					Description: "High-performance and ultra-fast model",
					ModelID:     "gemini-2.0-flash",
				},
				{
					Name:        "Gemini 2.5 Pro Exp 03-25",
					Description: "Experimental model with advanced features",
					ModelID:     "gemini-2.5-pro-exp-03-25",
				},
				{
					Name:        "Gemini 2.0 Flash Lite",
					Description: "Lightweight and efficient version of Gemini 2.0",
					ModelID:     "gemini-2.0-flash-lite",
				},
				{
					Name:        "Gemini 1.5 Flash",
					Description: "Reliable performance with fewer resources",
					ModelID:     "gemini-1.5-flash",
				},
				{
					Name:        "Gemini 1.5 Flash 8B",
					Description: "Optimized for 8 billion-parameter tasks",
					ModelID:     "gemini-1.5-flash-8b",
				},
				{
					Name:        "Gemini 1.5 Pro",
					Description: "Professional-grade model for large-scale applications",
					ModelID:     "gemini-1.5-pro",
				},
			},
		},
		{
			ID:          "openai",
			Name:        "OpenAI",
			Description: "OpenAI LLM provider",
			Models: []Model{
				{
					Name:        "GPT-4o Latest",
					Description: "Latest GPT-4o model",
					ModelID:     openai.ChatModelChatgpt4oLatest,
				},
				{
					Name:        "GPT-4o Mini",
					Description: "Optimized GPT-4o Mini model",
					ModelID:     openai.ChatModelGPT4oMini,
				},
				{
					Name:        "GPT-4",
					Description: "Standard GPT-4 model",
					ModelID:     openai.ChatModelGPT4,
				},
				{
					Name:        "GPT-4 Turbo",
					Description: "Most capable GPT-4 model for various tasks",
					ModelID:     openai.ChatModelGPT4Turbo,
				},
				{
					Name:        "GPT-3.5 Turbo",
					Description: "Efficient model balancing performance and speed",
					ModelID:     openai.ChatModelGPT3_5Turbo,
				},
				{
					Name:        "GPT-4.5 Preview",
					Description: "Last GPT-4.5 model from OpenAI",
					ModelID:     openai.ChatModelGPT4_5Preview,
				},
			},
		},
	}
}
