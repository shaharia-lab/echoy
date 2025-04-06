package tools

import (
	"encoding/json"
	"github.com/shaharia-lab/echoy/internal/api"
	"github.com/shaharia-lab/goai/mcp"
	"net/http"
)

type Provider struct {
	tools []Tool
}

// NewProvider creates a new instance of the Provider
func NewProvider(tools []mcp.Tool) *Provider {
	return &Provider{
		tools: convertToTools(tools),
	}
}

// ListTools lists all available tools
func (p *Provider) ListTools() ([]Tool, error) {
	return p.tools, nil
}

// ListToolsHTTPHandler handles HTTP requests to list all tools
func (p *Provider) ListToolsHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tools, err := p.ListTools()
		if err != nil {
			http.Error(w, "Failed to list tools", http.StatusInternalServerError)
			return
		}

		apiResponse := ToolListResponse{
			Tools: tools,
			Pagination: api.Pagination{
				Page:    1,
				PerPage: len(tools),
				Total:   len(tools),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResponse)
	}
}

// GetToolByName retrieves a tool by its name
func (p *Provider) GetToolByName(tools []Tool, name string) *Tool {
	for _, tool := range tools {
		if tool.Name == name {
			return &tool
		}
	}
	return nil
}

// GetToolByNameHTTPHandler handles HTTP requests to get a tool by its name
func (p *Provider) GetToolByNameHTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		tool := p.GetToolByName(p.tools, name)
		if tool == nil {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tool)
	}
}

func convertToTools(tools []mcp.Tool) []Tool {
	toolList := make([]Tool, len(tools))
	for i, tool := range tools {
		toolList[i] = Tool{
			Name:        tool.Name,
			Description: tool.Description,
		}
	}
	return toolList
}
