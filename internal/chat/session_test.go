package chat

import (
	"bufio"
	"context"
	"errors"
	chatMock "github.com/shaharia-lab/echoy/internal/chat/mocks"
	"github.com/shaharia-lab/echoy/internal/chat/types"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/echoy/internal/theme/mocks"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/goai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupMockTheme(t *testing.T) *mocks.MockTheme {
	mockTheme := mocks.NewMockTheme(t)
	mockWriter := mocks.NewMockWriter(t)

	mockTheme.EXPECT().Primary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Secondary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Info().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Success().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Warning().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Error().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Subtle().Return(mockWriter).Maybe()

	mockWriter.EXPECT().Print(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Println(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Printf(mock.Anything, mock.Anything).Return().Maybe()

	return mockTheme
}

func setupTestSession(t *testing.T) (*Session, *chatMock.MockService, *chatMock.MockHistoryService) {
	mockConfig := &config.Config{}
	mockTheme := setupMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()
	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			ch <- true
		},
	}

	return session, mockChatService, mockHistoryService
}

func TestNewChatSession(t *testing.T) {
	mockConfig := &config.Config{}
	mockTheme := setupMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()
	mockHistoryService.EXPECT().
		CreateChat(mock.Anything).
		Return(&goai.ChatHistory{SessionID: sessionUUID}, nil)

	session, err := NewChatSession(mockConfig, mockTheme, mockChatService, mockHistoryService)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionUUID, session.sessionID)
}

func TestNewChatSession_Error(t *testing.T) {
	mockConfig := &config.Config{}
	mockTheme := setupMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	expectedErr := errors.New("creation error")
	mockHistoryService.EXPECT().
		CreateChat(mock.Anything).
		Return(nil, expectedErr)

	session, err := NewChatSession(mockConfig, mockTheme, mockChatService, mockHistoryService)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestProcessMessage(t *testing.T) {
	session, mockChatService, _ := setupTestSession(t)
	session.thinkingAnimationFunc = func(theme theme.Theme, ch chan bool) {
		go func() {
			<-ch
		}()
	}

	ctx := context.Background()
	input := "test input"
	response := types.ChatResponse{
		ChatUUID:    uuid.UUID{}.String(),
		Answer:      "test response",
		InputToken:  0,
		OutputToken: 0,
	}

	mockChatService.EXPECT().
		Chat(ctx, session.sessionID, input).
		Return(response, nil)

	done := make(chan struct{})
	go func() {
		err := session.processMessage(ctx, input)
		assert.NoError(t, err)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestProcessMessage_Error(t *testing.T) {
	mockConfig := &config.Config{}
	mockTheme := mocks.NewMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()
	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			for range ch {
			}
		},
	}

	ctx := context.Background()
	input := "test input"
	expectedErr := errors.New("chat error")

	mockChatService.EXPECT().
		Chat(ctx, sessionUUID, input).
		Return(types.ChatResponse{}, expectedErr)

	err := session.processMessage(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestProcessMessageStreaming(t *testing.T) {
	mockConfig := &config.Config{}
	mockTheme := setupMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()

	var thinkingMutex sync.Mutex
	thinkingCalled := false

	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			thinkingMutex.Lock()
			thinkingCalled = true
			thinkingMutex.Unlock()

			for range ch {
			}
		},
	}

	ctx := context.Background()
	input := "test input"
	streamChan := make(chan goai.StreamingLLMResponse)

	mockChatService.EXPECT().
		ChatStreaming(ctx, sessionUUID, input).
		Return(streamChan, nil)

	errChan := make(chan error)
	go func() {
		errChan <- session.processMessageStreaming(ctx, input)
	}()

	go func() {
		time.Sleep(50 * time.Millisecond)

		streamChan <- goai.StreamingLLMResponse{
			Text: "test ",
			Done: false,
		}

		time.Sleep(50 * time.Millisecond)

		streamChan <- goai.StreamingLLMResponse{
			Text: "response",
			Done: true,
		}

		close(streamChan)
	}()

	select {
	case err := <-errChan:
		assert.NoError(t, err)

		thinkingMutex.Lock()
		called := thinkingCalled
		thinkingMutex.Unlock()

		assert.True(t, called, "Thinking animation function should be called")
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out")
	}
}

func TestProcessMessageStreaming_InitialError(t *testing.T) {
	mockConfig := &config.Config{}
	mockTheme := setupMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()

	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			select {
			case <-ch:
			case <-time.After(100 * time.Millisecond):
				return
			}
		},
	}

	ctx := context.Background()
	input := "test input"

	testErr := errors.New("test error")
	mockChatService.EXPECT().
		ChatStreaming(ctx, sessionUUID, input).
		Return(nil, testErr)

	err := session.processMessageStreaming(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), testErr.Error())
}

func TestProcessMessageStreaming_ResponseError(t *testing.T) {
	session, mockChatService, _ := setupTestSession(t)

	session.thinkingAnimationFunc = func(theme theme.Theme, ch chan bool) {
		go func() {
			<-ch
		}()
	}

	ctx := context.Background()
	input := "test input"
	expectedErr := errors.New("streaming response error")

	streamingChan := make(chan goai.StreamingLLMResponse)
	mockChatService.EXPECT().
		ChatStreaming(ctx, session.sessionID, input).
		Return(streamingChan, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		streamingChan <- goai.StreamingLLMResponse{
			Text: "part1 ",
			Done: false,
		}
		streamingChan <- goai.StreamingLLMResponse{
			Error: expectedErr,
		}
		close(streamingChan)
	}()

	err := session.processMessageStreaming(ctx, input)
	wg.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
}

func TestStart_StreamingEnabled(t *testing.T) {
	mockConfig := &config.Config{
		LLM: config.LLMConfig{
			Streaming: true,
		},
		User: config.UserConfig{
			Name: "Test User",
		},
	}
	mockTheme := mocks.NewMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()

	mockInput := "Hello\n\nexit\n"
	mockReader := bufio.NewReader(strings.NewReader(mockInput))

	mockWriter := mocks.NewMockWriter(t)
	mockTheme.EXPECT().Primary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Secondary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Info().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Success().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Warning().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Error().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Subtle().Return(mockWriter).Maybe()

	mockWriter.EXPECT().Println(mock.Anything, mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Println(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Print(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Printf(mock.Anything, mock.Anything).Return().Maybe()

	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		reader:             mockReader,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			go func() {
				select {
				case <-ch:
				case <-time.After(100 * time.Millisecond):
					return
				}
			}()
		},
	}

	ctx := context.Background()

	streamingChan := make(chan goai.StreamingLLMResponse)
	mockChatService.EXPECT().
		ChatStreaming(ctx, sessionUUID, "Hello").
		Return(streamingChan, nil)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond)
		streamingChan <- goai.StreamingLLMResponse{
			Text: "Test response",
			Done: true,
		}
		close(streamingChan)
	}()

	err := session.Start(ctx)

	wg.Wait()

	if err != nil && !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Unexpected error: %v", err)
	}
}
func TestStart_NoStreaming(t *testing.T) {
	mockConfig := &config.Config{
		LLM: config.LLMConfig{
			Streaming: false,
		},
		User: config.UserConfig{
			Name: "Test User",
		},
	}
	mockTheme := mocks.NewMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()

	mockInput := "Hello\n\nexit\n"
	mockReader := bufio.NewReader(strings.NewReader(mockInput))

	mockWriter := mocks.NewMockWriter(t)
	mockTheme.EXPECT().Primary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Secondary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Info().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Success().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Warning().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Error().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Subtle().Return(mockWriter).Maybe()

	mockWriter.EXPECT().Println(mock.Anything, mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Println(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Print(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Printf(mock.Anything, mock.Anything).Return().Maybe()

	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		reader:             mockReader,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			go func() {
				select {
				case <-ch:
				case <-time.After(100 * time.Millisecond):
					return
				}
			}()
		},
	}

	ctx := context.Background()

	mockChatService.EXPECT().
		Chat(ctx, sessionUUID, "Hello").
		Return(types.ChatResponse{Answer: "Response text"}, nil)

	err := session.Start(ctx)

	if err != nil && !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("Received expected EOF error: %v", err)
	}
}

func TestStart_ClearCommand(t *testing.T) {
	mockConfig := &config.Config{
		User: config.UserConfig{
			Name: "Test User",
		},
	}
	mockTheme := mocks.NewMockTheme(t)
	mockChatService := chatMock.NewMockService(t)
	mockHistoryService := chatMock.NewMockHistoryService(t)

	sessionUUID := uuid.New().String()

	mockInput := "clear\n\nexit\n"
	mockReader := bufio.NewReader(strings.NewReader(mockInput))

	mockWriter := mocks.NewMockWriter(t)
	mockTheme.EXPECT().Primary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Secondary().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Info().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Success().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Warning().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Error().Return(mockWriter).Maybe()
	mockTheme.EXPECT().Subtle().Return(mockWriter).Maybe()

	mockWriter.EXPECT().Println(mock.Anything, mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Println(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Print(mock.Anything).Return().Maybe()
	mockWriter.EXPECT().Printf(mock.Anything, mock.Anything).Return().Maybe()

	session := &Session{
		config:             mockConfig,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryService: mockHistoryService,
		sessionID:          sessionUUID,
		reader:             mockReader,
		thinkingAnimationFunc: func(theme theme.Theme, ch chan bool) {
			go func() {
				select {
				case <-ch:
				case <-time.After(100 * time.Millisecond):
					return
				}
			}()
		},
	}

	ctx := context.Background()

	err := session.Start(ctx)

	if err != nil && !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Unexpected error: %v", err)
	} else {
		t.Logf("Received expected EOF error: %v", err)
	}
}
