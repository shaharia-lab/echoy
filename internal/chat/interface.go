package chat

import (
	"context"
	"github.com/google/uuid"
	"github.com/shaharia-lab/goai"
)

// LLMService defines the interface for language model interactions
type LLMService interface {
	Generate(ctx context.Context, messages []goai.LLMMessage) (goai.LLMResponse, error)
	GenerateStream(ctx context.Context, messages []goai.LLMMessage) (<-chan goai.StreamingLLMResponse, error)
}

// ChatHistoryService defines operations for chat history management
type ChatHistoryService interface {
	CreateChat(ctx context.Context) (*goai.ChatHistory, error)
	AddMessage(ctx context.Context, uuid uuid.UUID, message goai.ChatHistoryMessage) error
	GetChat(ctx context.Context, uuid uuid.UUID) (*goai.ChatHistory, error)
}

// ChatService provides chat functionality using the LLM
type ChatService interface {
	Chat(ctx context.Context, sessionID uuid.UUID, message string) (goai.LLMResponse, error)
	ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error)
}
