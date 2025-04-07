package chat

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/chat/mocks"
	mocks2 "github.com/shaharia-lab/echoy/internal/llm/mocks"
	"github.com/shaharia-lab/goai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestServiceImpl_Chat(t *testing.T) {
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
			mockHistoryService := new(mocks.MockHistoryService)
			mockLLMService := new(mocks2.MockService)

			chatService := NewChatService(mockLLMService, mockHistoryService)

			ctx := context.Background()

			if tc.mockAddUserError != nil {
				mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
					return msg.Role == goai.UserRole && msg.Text == tc.userMessage
				})).Return(tc.mockAddUserError)
			} else {
				mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
					return msg.Role == goai.UserRole && msg.Text == tc.userMessage
				})).Return(nil)

				expectedLLMMessage := []goai.LLMMessage{{
					Role: goai.UserRole,
					Text: tc.userMessage,
				}}

				mockLLMService.On("Generate", ctx, expectedLLMMessage).Return(tc.mockLLMResponse, tc.mockLLMError)

				if tc.mockLLMError == nil {
					mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
						return msg.Role == goai.AssistantRole && msg.Text == tc.mockLLMResponse.Text
					})).Return(tc.mockAddAssistantError)
				}
			}

			response, err := chatService.Chat(ctx, tc.sessionID, tc.userMessage)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockLLMResponse.Text, response.Answer)
			}

			mockHistoryService.AssertExpectations(t)
			mockLLMService.AssertExpectations(t)
		})
	}
}

func TestServiceImpl_ChatStreaming(t *testing.T) {
	testCases := []struct {
		name             string
		sessionID        uuid.UUID
		userMessage      string
		mockAddUserError error
		mockStreamError  error
		expectedError    bool
	}{
		{
			name:            "successful chat streaming",
			sessionID:       uuid.New(),
			userMessage:     "Hello, how are you?",
			mockStreamError: nil,
		},
		{
			name:             "error adding user message",
			sessionID:        uuid.New(),
			userMessage:      "Hello",
			mockAddUserError: errors.New("failed to add user message"),
			expectedError:    true,
		},
		{
			name:            "error generating streaming response",
			sessionID:       uuid.New(),
			userMessage:     "Hello",
			mockStreamError: errors.New("streaming service error"),
			expectedError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockHistoryService := new(mocks.MockHistoryService)
			mockLLMService := new(mocks2.MockService)
			chatService := NewChatService(mockLLMService, mockHistoryService)

			ctx := context.Background()
			if tc.mockAddUserError != nil {
				mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
					return msg.Role == goai.UserRole && msg.Text == tc.userMessage
				})).Return(tc.mockAddUserError)
			} else {
				mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
					return msg.Role == goai.UserRole && msg.Text == tc.userMessage
				})).Return(nil)

				expectedLLMMessage := []goai.LLMMessage{{
					Role: goai.UserRole,
					Text: tc.userMessage,
				}}

				if tc.mockStreamError != nil {

					mockLLMService.On("GenerateStream", ctx, expectedLLMMessage).Return(nil, tc.mockStreamError)
				} else {

					mockRespChan := make(chan goai.StreamingLLMResponse)
					mockLLMService.On("GenerateStream", ctx, expectedLLMMessage).Return((<-chan goai.StreamingLLMResponse)(mockRespChan), nil)
					mockHistoryService.On("AddMessage", ctx, tc.sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
						return msg.Role == goai.AssistantRole && msg.Text == "Hello, world! How can I help?"
					})).Return(nil)

					resultChan, err := chatService.ChatStreaming(ctx, tc.sessionID, tc.userMessage)

					assert.NoError(t, err)
					assert.NotNil(t, resultChan)

					go func() {
						time.Sleep(50 * time.Millisecond)
						mockRespChan <- goai.StreamingLLMResponse{Text: "Hello", Error: nil, Done: false}
						time.Sleep(10 * time.Millisecond)
						mockRespChan <- goai.StreamingLLMResponse{Text: ", world!", Error: nil, Done: false}
						time.Sleep(10 * time.Millisecond)
						mockRespChan <- goai.StreamingLLMResponse{Text: " How can I help?", Error: nil, Done: true}
						close(mockRespChan)
					}()

					var responses []goai.StreamingLLMResponse
					timeout := time.After(1 * time.Second)

					for {
						select {
						case resp, ok := <-resultChan:
							if !ok {
								goto checkResponses
							}
							responses = append(responses, resp)
							t.Logf("Received response: %+v", resp)
						case <-timeout:
							t.Log("Timeout waiting for responses")
							goto checkResponses
						}
					}

				checkResponses:
					assert.Equal(t, 3, len(responses), "Should receive exactly 3 streaming responses")
					if len(responses) > 0 {
						assert.Equal(t, "Hello", responses[0].Text)
						assert.False(t, responses[0].Done)
					}

					if len(responses) > 1 {
						assert.Equal(t, ", world!", responses[1].Text)
						assert.False(t, responses[1].Done)
					}

					if len(responses) > 2 {
						assert.Equal(t, " How can I help?", responses[2].Text)
						assert.True(t, responses[2].Done)
					}

					return
				}
			}

			resultChan, err := chatService.ChatStreaming(ctx, tc.sessionID, tc.userMessage)
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, resultChan)
			}

			mockHistoryService.AssertExpectations(t)
			mockLLMService.AssertExpectations(t)
		})
	}
}

func TestProcessStreamingResponse(t *testing.T) {
	t.Run("successful processing", func(t *testing.T) {
		mockHistoryService := new(mocks.MockHistoryService)
		mockLLMService := new(mocks2.MockService)
		chatService := NewChatService(mockLLMService, mockHistoryService)

		ctx := context.Background()
		sessionID := uuid.New()

		respChan := make(chan goai.StreamingLLMResponse, 3)

		mockHistoryService.On("AddMessage", mock.Anything, sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
			return msg.Role == goai.AssistantRole
		})).Return(nil)

		go chatService.processStreamingResponse(ctx, sessionID, respChan)

		respChan <- goai.StreamingLLMResponse{Text: "Hello", Error: nil}
		respChan <- goai.StreamingLLMResponse{Text: ", world!", Error: nil}
		respChan <- goai.StreamingLLMResponse{Text: " This is a test.", Error: nil}
		close(respChan)

		time.Sleep(100 * time.Millisecond)

		mockHistoryService.AssertExpectations(t)
	})

	t.Run("processing with errors", func(t *testing.T) {
		mockHistoryService := new(mocks.MockHistoryService)
		mockLLMService := new(mocks2.MockService)
		chatService := NewChatService(mockLLMService, mockHistoryService)

		ctx := context.Background()
		sessionID := uuid.New()

		respChan := make(chan goai.StreamingLLMResponse, 4)

		mockHistoryService.On("AddMessage", ctx, sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
			return msg.Role == goai.AssistantRole
		})).Return(nil)

		go chatService.processStreamingResponse(ctx, sessionID, respChan)

		respChan <- goai.StreamingLLMResponse{Text: "Hello", Error: nil}
		respChan <- goai.StreamingLLMResponse{Text: ", world!", Error: errors.New("error in streaming")}
		respChan <- goai.StreamingLLMResponse{Text: " This is a test.", Error: nil}
		close(respChan)

		time.Sleep(100 * time.Millisecond)

		mockHistoryService.AssertExpectations(t)
	})

	t.Run("error saving to history", func(t *testing.T) {
		mockHistoryService := new(mocks.MockHistoryService)
		mockLLMService := new(mocks2.MockService)
		chatService := NewChatService(mockLLMService, mockHistoryService)

		ctx := context.Background()
		sessionID := uuid.New()

		respChan := make(chan goai.StreamingLLMResponse, 2)

		mockHistoryService.On("AddMessage", ctx, sessionID, mock.MatchedBy(func(msg goai.ChatHistoryMessage) bool {
			return msg.Role == goai.AssistantRole
		})).Return(errors.New("failed to save to history"))

		go chatService.processStreamingResponse(ctx, sessionID, respChan)

		respChan <- goai.StreamingLLMResponse{Text: "Hello", Error: nil}
		respChan <- goai.StreamingLLMResponse{Text: ", world!", Error: nil}
		close(respChan)

		time.Sleep(100 * time.Millisecond)

		mockHistoryService.AssertExpectations(t)
	})
}
