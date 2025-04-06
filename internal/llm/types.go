package llm

import "github.com/shaharia-lab/echoy/internal/api"

type Model struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelID     string `json:"modelId"`
}

// Provider represents a provider of large language models
type Provider struct {
	ID          string
	Name        string
	Description string
	Models      []Model
}

// ListProvidersResponse is the response structure for listing LLM providers
type ListProvidersResponse struct {
	Providers []Provider `json:"providers"`
	api.Pagination
}
