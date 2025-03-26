package chat

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/goai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
	"time"
)

// Mock implementations
type MockTheme struct {
	mock.Mock
}

func (m *MockTheme) Info() theme.Writer {
	args := m.Called()
	return args.Get(0).(theme.PrintWriter)
}

func (m *MockTheme) Subtle() theme.PrintWriter {
	args := m.Called()
	return args.Get(0).(theme.PrintWriter)
}

func (m *MockTheme) Secondary() theme.PrintWriter {
	args := m.Called()
	return args.Get(0).(theme.PrintWriter)
}

func (m *MockTheme) Primary() theme.PrintWriter {
	args := m.Called()
	return args.Get(0).(theme.PrintWriter)
}

func (m *MockTheme) Warning() theme.PrintWriter {
	args := m.Called()
	return args.Get(0).(theme.PrintWriter)
}

type MockPrintWriter struct {
	mock.Mock
	buffer *bytes.Buffer
}

func NewMockPrintWriter() *MockPrintWriter {
	return &MockPrintWriter{
		buffer: &bytes.Buffer{},
	}
}

func (m *MockPrintWriter) Println(a ...interface{}) (int, error) {
	m.Called(a)
	return fmt.Fprintln(m.buffer, a...)
}

func (m *MockPrintWriter) Printf(format string, a ...interface{}) (int, error) {
	m.Called(format, a)
	return fmt.Fprintf(m.buffer, format, a...)
}

func (m *MockPrintWriter) Print(a ...interface{}) (int, error) {
	m.Called(a)
	return fmt.Fprint(m.buffer, a...)
}

type MockChatHistoryStorage struct {
	mock.Mock
}

func (m *MockChatHistoryStorage) CreateChat(ctx context.Context) (*goai.ChatHistory, error) {
	args := m.Called(ctx)
	return args.Get(0).(*goai.ChatSession), args.Error(1)
}

func (m *MockChatHistoryStorage) GetChat(ctx context.Context, chatID uuid.UUID) (*goai.ChatSession, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*goai.ChatSession), args.Error(1)
}

func (m *MockChatHistoryStorage) AddMessage(ctx context.Context, chatID uuid.UUID, message goai.ChatMessage) error {
	args := m.Called(ctx, chatID, message)
	return args.Error(0)
}

func (m *MockChatHistoryStorage) GetMessages(ctx context.Context, chatID uuid.UUID) ([]goai.ChatMessage, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]goai.ChatMessage), args.Error(1)
}

type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) Chat(ctx context.Context, sessionID uuid.UUID, input string) (*goai.ChatResponse, error) {
	args := m.Called(ctx, sessionID, input)
	return args.Get(0).(*goai.ChatResponse), args.Error(1)
}

// Test setup helper
func setupSessionTest() (*config.Config, *MockTheme, *MockChatService, *MockChatHistoryStorage, *MockPrintWriter) {
	cfg := &config.Config{
		User: config.UserConfig{
			Name: "TestUser",
		},
	}

	mockInfo := NewMockPrintWriter()
	mockInfo.On("Println", mock.Anything).Return(0, nil)
	mockInfo.On("Print", mock.Anything).Return(0, nil)
	mockInfo.On("Printf", mock.Anything, mock.Anything).Return(0, nil)

	mockSubtle := NewMockPrintWriter()
	mockSubtle.On("Println", mock.Anything).Return(0, nil)
	mockSubtle.On("Printf", mock.Anything, mock.Anything).Return(0, nil)
	mockSubtle.On("Print", mock.Anything).Return(0, nil)

	mockSecondary := NewMockPrintWriter()
	mockSecondary.On("Println", mock.Anything).Return(0, nil)
	mockSecondary.On("Printf", mock.Anything, mock.Anything).Return(0, nil)
	mockSecondary.On("Print", mock.Anything).Return(0, nil)

	mockPrimary := NewMockPrintWriter()
	mockPrimary.On("Println", mock.Anything).Return(0, nil)
	mockPrimary.On("Printf", mock.Anything, mock.Anything).Return(0, nil)
	mockPrimary.On("Print", mock.Anything).Return(0, nil)

	mockWarning := NewMockPrintWriter()
	mockWarning.On("Println", mock.Anything).Return(0, nil)
	mockWarning.On("Printf", mock.Anything, mock.Anything).Return(0, nil)
	mockWarning.On("Print", mock.Anything).Return(0, nil)

	mockTheme := new(MockTheme)
	mockTheme.On("Info").Return(mockInfo)
	mockTheme.On("Subtle").Return(mockSubtle)
	mockTheme.On("Secondary").Return(mockSecondary)
	mockTheme.On("Primary").Return(mockPrimary)
	mockTheme.On("Warning").Return(mockWarning)

	mockChatService := new(MockChatService)
	mockChatHistoryStorage := new(MockChatHistoryStorage)

	return cfg, mockTheme, mockChatService, mockChatHistoryStorage, mockPrimary
}

func TestNewChatSession(t *testing.T) {
	ctx := context.Background()
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	sessionID := uuid.New()
	mockChatHistoryStorage.On("CreateChat", ctx).Return(&goai.ChatSession{
		UUID: sessionID,
	}, nil)

	session, err := NewChatSession(cfg, mockTheme, mockChatService, mockChatHistoryStorage)

	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.sessionID)
	assert.Equal(t, cfg, session.config)
	mockChatHistoryStorage.AssertExpectations(t)
}

func TestNewChatSessionError(t *testing.T) {
	ctx := context.Background()
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	expectedErr := fmt.Errorf("storage error")
	mockChatHistoryStorage.On("CreateChat", ctx).Return(&goai.ChatSession{}, expectedErr)

	session, err := NewChatSession(cfg, mockTheme, mockChatService, mockChatHistoryStorage)

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, err.Error(), expectedErr.Error())
	mockChatHistoryStorage.AssertExpectations(t)
}

func TestShowWelcomeMessage(t *testing.T) {
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	session := &Session{
		config:             cfg,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryStorage: mockChatHistoryStorage,
		sessionID:          uuid.New(),
	}

	session.showWelcomeMessage()

	mockTheme.AssertExpectations(t)
}

func TestReadUserInput(t *testing.T) {
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	t.Run("Single line input", func(t *testing.T) {
		inputStr := "Hello\n\n"
		reader := bufio.NewReader(strings.NewReader(inputStr))

		session := &Session{
			config:             cfg,
			theme:              mockTheme,
			chatService:        mockChatService,
			chatHistoryStorage: mockChatHistoryStorage,
			sessionID:          uuid.New(),
			reader:             reader,
		}

		result, err := session.readUserInput()

		assert.NoError(t, err)
		assert.Equal(t, "Hello", result)
	})

	t.Run("Multi-line input", func(t *testing.T) {
		inputStr := "Line 1\nLine 2\nLine 3\n\n"
		reader := bufio.NewReader(strings.NewReader(inputStr))

		session := &Session{
			config:             cfg,
			theme:              mockTheme,
			chatService:        mockChatService,
			chatHistoryStorage: mockChatHistoryStorage,
			sessionID:          uuid.New(),
			reader:             reader,
		}

		result, err := session.readUserInput()

		assert.NoError(t, err)
		assert.Equal(t, "Line 1\nLine 2\nLine 3", result)
	})
}

func TestProcessMessage(t *testing.T) {
	ctx := context.Background()
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	sessionID := uuid.New()
	input := "Hello AI"
	expectedResponse := &goai.ChatResponse{
		Text: "Hello human",
	}

	mockChatService.On("Chat", ctx, sessionID, input).Return(expectedResponse, nil)

	session := &Session{
		config:             cfg,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryStorage: mockChatHistoryStorage,
		sessionID:          sessionID,
	}

	err := session.processMessage(ctx, input)

	assert.NoError(t, err)
	mockChatService.AssertExpectations(t)
}

func TestProcessMessageError(t *testing.T) {
	ctx := context.Background()
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	sessionID := uuid.New()
	input := "Hello AI"
	expectedErr := fmt.Errorf("service error")

	mockChatService.On("Chat", ctx, sessionID, input).Return(&goai.ChatResponse{}, expectedErr)

	session := &Session{
		config:             cfg,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryStorage: mockChatHistoryStorage,
		sessionID:          sessionID,
	}

	err := session.processMessage(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedErr.Error())
	mockChatService.AssertExpectations(t)
}

func TestShowThinkingAnimation(t *testing.T) {
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	session := &Session{
		config:             cfg,
		theme:              mockTheme,
		chatService:        mockChatService,
		chatHistoryStorage: mockChatHistoryStorage,
		sessionID:          uuid.New(),
	}

	thinking := make(chan bool)
	go func() {
		// Let animation run briefly
		time.Sleep(500 * time.Millisecond)
		thinking <- true
	}()

	session.showThinkingAnimation(thinking)
	// If we get here without deadlock, the test passes
}

func TestStart(t *testing.T) {
	ctx := context.Background()
	cfg, mockTheme, mockChatService, mockChatHistoryStorage, _ := setupSessionTest()

	t.Run("Exit command", func(t *testing.T) {
		inputStr := "exit\n"
		reader := bufio.NewReader(strings.NewReader(inputStr))

		session := &Session{
			config:             cfg,
			theme:              mockTheme,
			chatService:        mockChatService,
			chatHistoryStorage: mockChatHistoryStorage,
			sessionID:          uuid.New(),
			reader:             reader,
		}

		err := session.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("Clear command", func(t *testing.T) {
		inputStr := "clear\nexit\n"
		reader := bufio.NewReader(strings.NewReader(inputStr))

		session := &Session{
			config:             cfg,
			theme:              mockTheme,
			chatService:        mockChatService,
			chatHistoryStorage: mockChatHistoryStorage,
			sessionID:          uuid.New(),
			reader:             reader,
		}

		err := session.Start(ctx)
		assert.NoError(t, err)
	})

	t.Run("Process message", func(t *testing.T) {
		inputStr := "hello\n\nexit\n"
		reader := bufio.NewReader(strings.NewReader(inputStr))
		sessionID := uuid.New()

		mockChatService.On("Chat", ctx, sessionID, "hello").Return(&goai.ChatResponse{
			Text: "Hi there!",
		}, nil)

		session := &Session{
			config:             cfg,
			theme:              mockTheme,
			chatService:        mockChatService,
			chatHistoryStorage: mockChatHistoryStorage,
			sessionID:          sessionID,
			reader:             reader,
		}

		err := session.Start(ctx)
		assert.NoError(t, err)
		mockChatService.AssertExpectations(t)
	})
}
