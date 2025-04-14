package types

import (
	"github.com/shaharia-lab/echoy/internal/api"
	"github.com/shaharia-lab/goai"
)

type ModelSettings struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int64   `json:"maxTokens"`
	TopP        float64 `json:"topP"`
	TopK        int64   `json:"topK"`
}

type LLMProvider struct {
	Provider string `json:"provider"`
	ModelID  string `json:"modelId"`
}

type StreamSettings struct {
	ChunkSize int `json:"chunk_size"`
	DelayMs   int `json:"delay_ms"`
}

type ChatRequest struct {
	ChatUUID       string         `json:"chat_uuid"`
	Question       string         `json:"question"`
	SelectedTools  []string       `json:"selectedTools"`
	ModelSettings  ModelSettings  `json:"modelSettings"`
	LLMProvider    LLMProvider    `json:"llmProvider"`
	StreamSettings StreamSettings `json:"stream_settings"`
}

type ChatResponse struct {
	ChatUUID    string `json:"chat_uuid"`
	Answer      string `json:"answer"`
	InputToken  int    `json:"input_token"`
	OutputToken int    `json:"output_token"`
}

type ChatHistoryList struct {
	Chats []goai.ChatHistory `json:"chats"`
	api.Pagination
}
