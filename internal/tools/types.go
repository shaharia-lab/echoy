package tools

import "github.com/shaharia-lab/echoy/internal/api"

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ToolListResponse struct {
	Tools []Tool `json:"tools"`
	api.Pagination
}
