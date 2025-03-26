package chat

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/goai"
	"time"
)

// ChatServiceImpl implements the ChatService interface
type ChatServiceImpl struct {
	llmService     LLMService
	historyService ChatHistoryService
}

// NewChatService creates a new chat service
func NewChatService(llmService LLMService, historyService ChatHistoryService) *ChatServiceImpl {
	return &ChatServiceImpl{
		llmService:     llmService,
		historyService: historyService,
	}
}

// Chat provides non-streaming chat functionality
func (s *ChatServiceImpl) Chat(ctx context.Context, sessionID uuid.UUID, message string) (goai.LLMResponse, error) {
	// Create user message
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: message,
	}

	// Save user message to history
	if err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage:  userMessage,
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to add message to chat history: %w", err)
	}

	// Generate response
	llmResponse, err := s.llmService.Generate(ctx, []goai.LLMMessage{userMessage})
	if err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to generate response: %w", err)
	}

	// Save assistant response to history
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
func (s *ChatServiceImpl) ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error) {
	// Create user message
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: message,
	}

	// Save user message to history
	if err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage:  userMessage,
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("failed to add message to chat history: %w", err)
	}

	// Generate streaming response
	responseChan, err := s.llmService.GenerateStream(ctx, []goai.LLMMessage{userMessage})
	if err != nil {
		return nil, fmt.Errorf("failed to generate streaming response: %w", err)
	}

	// Process the streaming response and save the complete response to history
	go s.processStreamingResponse(ctx, sessionID, responseChan)

	return responseChan, nil
}

// processStreamingResponse collects the streaming response and saves it to history
func (s *ChatServiceImpl) processStreamingResponse(ctx context.Context, sessionID uuid.UUID, responseChan <-chan goai.StreamingLLMResponse) {
	var completeResponse string

	for streamingResp := range responseChan {
		if streamingResp.Error != nil {
			// Log error but continue collection
			continue
		}
		completeResponse += streamingResp.Text
	}

	// Save complete response to history
	err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage: goai.LLMMessage{
			Role: goai.AssistantRole,
			Text: completeResponse,
		},
		GeneratedAt: time.Now().UTC(),
	})
	if err != nil {
		// Log the error since we can't return it from a goroutine
		fmt.Printf("Failed to save complete streaming response: %v\n", err)
	}
}
