package chat

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/chat/mocks"
	"github.com/shaharia-lab/goai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestServiceImpl_Chat(t *testing.T) {
	// Test cases
	testCases := []struct {
		name                  string
		sessionID             uuid.UUID
		userMessage           string
		mockLLMResponse       goai.LLMResponse
		mockLLMError          error
		mockAddUserError      error
		mockAddAssistantError error
		expectedError         bool
	}{
		{
			name:            "successful chat",
			sessionID:       uuid.New(),
			userMessage:     "Hello, how are you?",
			mockLLMResponse: goai.LLMResponse{Text: "I'm doing well, thank you!"},
			mockLLMError:    nil,
		},
		{
			name:             "error adding user message",
			sessionID:        uuid.New(),
			userMessage:      "Hello",
			mockAddUserError: errors.New("failed to add user message"),
			expectedError:    true,
		},
		{
			name:          "error generating LLM response",
			sessionID:     uuid.New(),
			userMessage:   "Hello",
			mockLLMError:  errors.New("LLM service error"),
			expectedError: true,
		},
		{
			name:                  "error adding assistant response",
			sessionID:             uuid.New(),
			userMessage:           "Hello",
			mockLLMResponse:       goai.LLMResponse{Text: "Response"},
			mockAddAssistantError: errors.New("failed to add assistant message"),
			expectedError:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockHistoryService := new(mocks.MockHistoryService)
			mockLLMService := new(mocks.MockLLMService)

			// Create the service with mocks
			chatService := NewChatService(mockLLMService, mockHistoryService)

			ctx := context.Background()

			// Mock the AddMessage call for the user message
			if tc.mockAddUserError != nil {
				mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
					return msg.Role == goai.UserRole && msg.Text == tc.userMessage
				})).Return(tc.mockAddUserError)
			} else {
				mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
					return msg.Role == goai.UserRole && msg.Text == tc.userMessage
				})).Return(nil)

				// Only set up LLM expectations if the first AddMessage doesn't error
				expectedLLMMessage := []goai.LLMMessage{{
					Role: goai.UserRole,
					Text: tc.userMessage,
				}}

				mockLLMService.On("Generate", ctx, expectedLLMMessage).Return(tc.mockLLMResponse, tc.mockLLMError)

				// Only set up assistant message expectations if LLM doesn't error
				if tc.mockLLMError == nil {
					mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
						return msg.Role == goai.AssistantRole && msg.Text == tc.mockLLMResponse.Text
					})).Return(tc.mockAddAssistantError)
				}
			}

			// Call the method under test
			response, err := chatService.Chat(ctx, tc.sessionID, tc.userMessage)

			// Assertions
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockLLMResponse, response)
			}

			// Verify all expectations were met
			mockHistoryService.AssertExpectations(t)
			mockLLMService.AssertExpectations(t)
		})
	}
}
