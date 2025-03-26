package chat

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/goai"
	"time"
)

// ServiceImpl implements the ChatService interface
type ServiceImpl struct {
	llmService     LLMService
	historyService HistoryService
}

// NewChatService creates a new chat service
func NewChatService(llmService LLMService, historyService HistoryService) *ServiceImpl {
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

	responseChan, err := s.llmService.GenerateStream(ctx, []goai.LLMMessage{userMessage})
	if err != nil {
		return nil, fmt.Errorf("failed to generate streaming response: %w", err)
	}

	go s.processStreamingResponse(ctx, sessionID, responseChan)

	return responseChan, nil
}

// processStreamingResponse collects the streaming response and saves it to history
func (s *ServiceImpl) processStreamingResponse(ctx context.Context, sessionID uuid.UUID, responseChan <-chan goai.StreamingLLMResponse) {
	var completeResponse string

	for streamingResp := range responseChan {
		if streamingResp.Error != nil {
			continue
		}
		completeResponse += streamingResp.Text
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
