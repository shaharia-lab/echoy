package llm

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/goai"
	"strings"
)

// Service defines the interface for language model interactions
type Service interface {
	Generate(ctx context.Context, messages []goai.LLMMessage) (goai.LLMResponse, error)
	GenerateStream(ctx context.Context, messages []goai.LLMMessage) (<-chan goai.StreamingLLMResponse, error)
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	provider goai.LLMProvider
	config   goai.LLMRequestConfig
}

// NewLLMService creates a new LLM service
func NewLLMService(llmConfig config.LLMConfig) (*ServiceImpl, error) {
	provider, err := buildLLMProvider(llmConfig)
	if err != nil {
		return nil, err
	}
	toolsProvider := goai.NewToolsProvider()
	cfg := goai.NewRequestConfig(
		goai.WithMaxToken(1000),
		goai.WithTemperature(0.7),
		goai.UseToolsProvider(toolsProvider),
	)

	return &ServiceImpl{
		provider: provider,
		config:   cfg,
	}, nil
}

// Generate implements the Service interface
func (s *ServiceImpl) Generate(ctx context.Context, messages []goai.LLMMessage) (goai.LLMResponse, error) {
	llm := goai.NewLLMRequest(s.config, s.provider)
	return llm.Generate(ctx, messages)
}

// GenerateStream implements the Service interface
func (s *ServiceImpl) GenerateStream(ctx context.Context, messages []goai.LLMMessage) (<-chan goai.StreamingLLMResponse, error) {
	llm := goai.NewLLMRequest(s.config, s.provider)
	return llm.GenerateStream(ctx, messages)
}

// buildLLMProvider creates the appropriate LLM provider based on config
func buildLLMProvider(llmConfig config.LLMConfig) (goai.LLMProvider, error) {
	if llmConfig.Provider == "" {
		return nil, fmt.Errorf("llm provider not specified")
	}

	if llmConfig.Token == "" {
		return nil, fmt.Errorf("token for LLM provider not specified")
	}

	switch strings.ToLower(llmConfig.Provider) {
	case "anthropic":
		return goai.NewAnthropicLLMProvider(goai.AnthropicProviderConfig{
			Client: goai.NewAnthropicClient(llmConfig.Token),
			Model:  llmConfig.Model,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", llmConfig.Provider)
	}
}
