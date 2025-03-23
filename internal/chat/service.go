package chat

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/goai"
	"strings"
	"time"
)

type Service struct {
	llmProvider        goai.LLMProvider
	llmConfig          config.LLMConfig
	chatHistoryStorage goai.ChatHistoryStorage
	llmRequestConfig   goai.LLMRequestConfig
}

func NewChatService(llmConfig config.LLMConfig, chatHistoryStorage goai.ChatHistoryStorage) (*Service, error) {
	llmProvider, err := buildLLMProvider(llmConfig)
	if err != nil {
		return nil, err
	}

	llmRequestConfig := goai.LLMRequestConfig{
		MaxToken:       500,
		TopP:           0.5,
		Temperature:    0.5,
		TopK:           40,
		DisableTracing: true,
		AllowedTools:   []string{},
	}

	return &Service{
		llmConfig:          llmConfig,
		llmProvider:        llmProvider,
		chatHistoryStorage: chatHistoryStorage,
		llmRequestConfig:   llmRequestConfig,
	}, nil
}

func buildLLMProvider(llmConfig config.LLMConfig) (goai.LLMProvider, error) {
	var llmProvider goai.LLMProvider

	if llmConfig.Provider == "" {
		return llmProvider, fmt.Errorf("llm provider not specified")
	}

	if llmConfig.Token == "" {
		return llmProvider, fmt.Errorf("token for LLM provider not specified")
	}

	// Initialize the LLM provider
	switch strings.ToLower(llmConfig.Provider) {
	case "anthropic":
		llmProvider = goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
			Client: goai.NewAnthropicClient(llmConfig.Token),
			Model:  llmConfig.Model,
		})
	default:
		return llmProvider, fmt.Errorf("un-supported LLM provider: %s", llmConfig.Provider)
	}

	return llmProvider, nil
}

func (c *Service) Chat(ctx context.Context, sessionID uuid.UUID, message string) (goai.LLMResponse, error) {
	messages := []goai.LLMMessage{
		{
			Role: goai.UserRole,
			Text: message,
		},
	}

	for _, m := range messages {
		if err := c.chatHistoryStorage.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
			LLMMessage:  m,
			GeneratedAt: time.Now().UTC(),
		}); err != nil {
			return goai.LLMResponse{}, fmt.Errorf("failed to add message to chat history: %w", err)
		}
	}

	llm := goai.NewLLMRequest(goai.NewRequestConfig(
		goai.WithMaxToken(100),
		goai.WithTemperature(0.7),
		goai.UseToolsProvider(goai.NewToolsProvider()),
	), c.llmProvider)

	llmResponse, err := llm.Generate(ctx, messages)
	if err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to generate response: %w", err)
	}

	err = c.chatHistoryStorage.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
		LLMMessage: goai.LLMMessage{
			Role: goai.AssistantRole,
			Text: llmResponse.Text,
		},
		GeneratedAt: time.Now().UTC(),
	})
	if err != nil {
		return goai.LLMResponse{}, fmt.Errorf("failed to add message to chat history: %w", err)
	}

	return llmResponse, nil
}

func (c *Service) ChatStreaming(ctx context.Context, sessionID uuid.UUID, message string) (<-chan goai.StreamingLLMResponse, error) {
	messages := []goai.LLMMessage{
		{
			Role: goai.UserRole,
			Text: message,
		},
	}

	for _, m := range messages {
		if err := c.chatHistoryStorage.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
			LLMMessage:  m,
			GeneratedAt: time.Now().UTC(),
		}); err != nil {
			return nil, fmt.Errorf("failed to add message to chat history: %w", err)
		}
	}

	llm := goai.NewLLMRequest(c.llmRequestConfig, c.llmProvider)
	responseChan, err := llm.GenerateStream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	go func() {
		var completeResponse string
		for streamingResp := range responseChan {
			if streamingResp.Error != nil {
				// Handle errors in streaming response if necessary
				continue
			}
			completeResponse += streamingResp.Text
		}

		if err := c.chatHistoryStorage.AddMessage(ctx, sessionID, goai.ChatHistoryMessage{
			LLMMessage: goai.LLMMessage{
				Role: goai.AssistantRole,
				Text: completeResponse,
			},
			GeneratedAt: time.Now().UTC(),
		}); err != nil {
			// Handle error if necessary
		}
	}()

	return responseChan, nil
}
