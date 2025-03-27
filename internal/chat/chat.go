package chat

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/llm"
	"github.com/shaharia-lab/goai"
	"time"
)

// HistoryService defines operations for chat history management
type HistoryService interface {
	CreateChat(ctx context.Context) (*goai.ChatHistory, error)
	AddMessage(ctx context.Context, uuid uuid.UUID, message goai.ChatHistoryMessage) error
	GetChat(ctx context.Context, uuid uuid.UUID) (*goai.ChatHistory, error)
}

// Service provides chat functionality using the LLM
type Service interface {
	Chat(ctx context.Context, sessionID uuid.UUID, message string) (goai.LLMResponse, error)
	ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error)
}

// ServiceImpl implements the ChatService interface
type ServiceImpl struct {
	llmService     llm.Service
	historyService HistoryService
}

// NewChatService creates a new chat service
func NewChatService(llmService llm.Service, historyService HistoryService) *ServiceImpl {
	return &ServiceImpl{
		llmService:     llmService,
		historyService: historyService,
	}
}

// Chat provides non-streaming chat functionality
func (s *ServiceImpl) Chat(ctx context.Context, sessionID uuid.UUID, message string) (goai.LLMResponse, error) {
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: message,
	}

	if err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage:  userMessage,
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to add message to chat history: %w", err)
	}

	llmResponse, err := s.llmService.Generate(ctx, []goai.LLMMessage{userMessage})
	if err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to generate response: %w", err)
	}

	err = s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage: goai.LLMMessage{
			Role: goai.AssistantRole,
			Text: llmResponse.Text,
		},
		GeneratedAt: time.Now().UTC(),
	})
	if err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to add response to chat history: %w", err)
	}

	return llmResponse, nil
}

// ChatStreaming provides streaming chat functionality
func (s *ServiceImpl) ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error) {
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: message,
	}

	if err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage:  userMessage,
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("failed to add message to chat history: %w", err)
	}

	sourceChan, err := s.llmService.GenerateStream(ctx, []goai.LLMMessage{userMessage})
	if err != nil {
		return nil, fmt.Errorf("failed to generate streaming response: %w", err)
	}

	// Create a new channel to broadcast responses
	resultChan := make(chan goai.StreamingLLMResponse)

	go func() {
		defer close(resultChan)

		var completeResponse string

		for streamingResp := range sourceChan {
			// Forward each response to our result channel
			select {
			case resultChan <- streamingResp:
				// Message forwarded
			case <-ctx.Done():
				return
			}

			// Process for history
			if streamingResp.Error != nil {
				continue
			}

			completeResponse += streamingResp.Text

			if streamingResp.Done {
				// Save complete response to history
				err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
					LLMMessage: goai.LLMMessage{
						Role: goai.AssistantRole,
						Text: completeResponse,
					},
					GeneratedAt: time.Now().UTC(),
				})
				if err != nil {
					fmt.Printf("Failed to save complete streaming response: %v\n", err)
				}
			}
		}
	}()

	return resultChan, nil
}

// processStreamingResponse collects the streaming response and saves it to history
func (s *ServiceImpl) processStreamingResponse(ctx context.Context, sessionID uuid.UUID, responseChan <-chan goai.StreamingLLMResponse) {
	var completeResponse string

	for streamingResp := range responseChan {
		if streamingResp.Error != nil {
			continue
		}

		completeResponse += streamingResp.Text

		if streamingResp.Done {
			break
		}
	}

	err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage: goai.LLMMessage{
			Role: goai.AssistantRole,
			Text: completeResponse,
		},
		GeneratedAt: time.Now().UTC(),
	})
	if err != nil {
		fmt.Printf("Failed to save complete streaming response: %v\n", err)
	}
}
