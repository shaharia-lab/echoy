package llm

import (
	"encoding/json"
	"github.com/shaharia-lab/echoy/internal/api"
	"net/http"
)

type LLMHandler struct {
	providers []Provider
}

func NewLLMHandler(providers []Provider) *LLMHandler {
	return &LLMHandler{
		providers: providers,
	}
}

// ListProvidersHTTPHandler handles HTTP requests to list all LLM providers
func (h *LLMHandler) ListProvidersHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the response header to JSON
		w.Header().Set("Content-Type", "application/json")

		// Create a response struct
		response := ListProvidersResponse{
			Providers: h.providers,
			Pagination: api.Pagination{
				Page:    1,
				PerPage: len(h.providers),
				Total:   len(h.providers),
			},
		}

		// Write the JSON response
		json.NewEncoder(w).Encode(response)
	}
}

func GetProviderByID(providers []Provider, id string) *Provider {
	for _, provider := range providers {
		if provider.ID == id {
			return &provider
		}
	}
	return nil
}

// GetProviderByIDHTTPHandler handles HTTP requests to get a specific LLM provider by ID
func (h *LLMHandler) GetProviderByIDHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the provider ID from the URL
		providerID := r.URL.Query().Get("id")

		// Find the provider by ID
		provider := GetProviderByID(h.providers, providerID)
		if provider == nil {
			http.Error(w, "Provider not found", http.StatusNotFound)
			return
		}

		// Set the response header to JSON
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON response
		json.NewEncoder(w).Encode(provider)
	}
}
