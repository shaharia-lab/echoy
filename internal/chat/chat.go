package chat

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/api"
	"github.com/shaharia-lab/echoy/internal/llm"
	"github.com/shaharia-lab/goai"
	"time"
)

// HistoryService defines operations for chat history management
type HistoryService interface {
	CreateChat(ctx context.Context) (*goai.ChatHistory, error)
	AddMessage(ctx context.Context, uuid uuid.UUID, message goai.ChatHistoryMessage) error
	GetChat(ctx context.Context, uuid uuid.UUID) (*goai.ChatHistory, error)
	ListChatHistories(ctx context.Context) ([]goai.ChatHistory, error)
}

// Service provides chat functionality using the LLM
type Service interface {
	Chat(ctx context.Context, sessionID uuid.UUID, message string) (ChatResponse, error)
	ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error)
	GetChatHistory(ctx context.Context, chatUUID uuid.UUID) (*goai.ChatHistory, error)
	GetListChatHistories(ctx context.Context) (ChatHistoryList, error)
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
func (s *ServiceImpl) Chat(ctx context.Context, sessionID uuid.UUID, message string) (ChatResponse, error) {
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: message,
	}

	if sessionID == uuid.Nil {
		chatHistory, err := s.historyService.CreateChat(ctx)
		if err != nil {
			return ChatResponse{}, fmt.Errorf("failed to create chat session: %w", err)
		}

		sessionID = chatHistory.UUID
	}

	if err := s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage:  userMessage,
		GeneratedAt: time.Now().UTC(),
	}); err != nil {
		return ChatResponse{}, fmt.Errorf("failed to add message to chat history: %w", err)
	}

	llmResponse, err := s.llmService.Generate(ctx, []goai.LLMMessage{userMessage})
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to generate response: %w", err)
	}

	err = s.historyService.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage: goai.LLMMessage{
			Role: goai.AssistantRole,
			Text: llmResponse.Text,
		},
		GeneratedAt: time.Now().UTC(),
	})
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to add response to chat history: %w", err)
	}

	return ChatResponse{
		ChatUUID:    sessionID,
		Answer:      llmResponse.Text,
		InputToken:  llmResponse.TotalInputToken,
		OutputToken: llmResponse.TotalOutputToken,
	}, nil
}

// GetChatHistory retrieves chat history for a given chat session
func (s *ServiceImpl) GetChatHistory(ctx context.Context, chatUUID uuid.UUID) (*goai.ChatHistory, error) {
	chatHistory, err := s.historyService.GetChat(ctx, chatUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat histories: %w", err)
	}

	return chatHistory, nil
}

// GetListChatHistories retrieves all chat histories
func (s *ServiceImpl) GetListChatHistories(ctx context.Context) (ChatHistoryList, error) {
	chatHistories, err := s.historyService.ListChatHistories(ctx)
	if err != nil {
		return ChatHistoryList{}, fmt.Errorf("failed to list chat histories: %w", err)
	}

	return ChatHistoryList{
		Chats: chatHistories,
		Pagination: api.Pagination{
			Page:    1,
			PerPage: len(chatHistories),
			Total:   len(chatHistories),
		},
	}, nil
}

// ChatStreaming provides streaming chat functionality
func (s *ServiceImpl) ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error) {
	userMessage := goai.LLMMessage{
		Role: goai.UserRole,
		Text: message,
	}

	if sessionID == uuid.Nil {
		chatHistory, err := s.historyService.CreateChat(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create chat session: %w", err)
		}
		sessionID = chatHistory.UUID
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
