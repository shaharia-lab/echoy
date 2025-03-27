package llm

import (
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_buildLLMProvider(t *testing.T) {
	tests := []struct {
		name       string
		config     config.LLMConfig
		wantErr    bool
		errMessage string
	}{
		{
			name: "valid anthropic provider",
			config: config.LLMConfig{
				Provider: "anthropic",
				Token:    "test-token",
				Model:    "claude-3",
			},
			wantErr: false,
		},
		{
			name: "empty provider",
			config: config.LLMConfig{
				Provider: "",
				Token:    "test-token",
			},
			wantErr:    true,
			errMessage: "llm provider not specified",
		},
		{
			name: "empty token",
			config: config.LLMConfig{
				Provider: "anthropic",
				Token:    "",
			},
			wantErr:    true,
			errMessage: "token for LLM provider not specified",
		},
		{
			name: "unsupported provider",
			config: config.LLMConfig{
				Provider: "unsupported",
				Token:    "test-token",
			},
			wantErr:    true,
			errMessage: "unsupported LLM provider: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := buildLLMProvider(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, provider)
				if tt.errMessage != "" {
					assert.Equal(t, tt.errMessage, err.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}
