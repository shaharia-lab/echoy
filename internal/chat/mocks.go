package chat

import (
	"context"
	"errors"
	"github.com/shaharia-lab/goai"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MockChatHistoryStorage implements ChatHistoryStorage interface for testing purposes
type MockChatHistoryStorage struct {
	mutex    sync.RWMutex
	chatLogs map[uuid.UUID]*goai.ChatHistory

	// Error simulation fields for testing error scenarios
	createChatErr        error
	addMessageErr        error
	getChatErr           error
	listChatHistoriesErr error
	deleteChatErr        error
}

// NewMockChatHistoryStorage creates a new mock storage implementation
func NewMockChatHistoryStorage() *MockChatHistoryStorage {
	return &MockChatHistoryStorage{
		chatLogs: make(map[uuid.UUID]*goai.ChatHistory),
	}
}

// CreateChat initializes a new chat conversation
func (m *MockChatHistoryStorage) CreateChat(_ context.Context) (*goai.ChatHistory, error) {
	if m.createChatErr != nil {
		return nil, m.createChatErr
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	chat := &goai.ChatHistory{
		UUID:      uuid.New(),
		Messages:  []goai.ChatHistoryMessage{},
		CreatedAt: time.Now(),
	}

	m.chatLogs[chat.UUID] = chat
	return chat, nil
}

// AddMessage adds a new message to an existing conversation
func (m *MockChatHistoryStorage) AddMessage(_ context.Context, uuid uuid.UUID, message goai.ChatHistoryMessage) error {
	if m.addMessageErr != nil {
		return m.addMessageErr
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	chat, exists := m.chatLogs[uuid]
	if !exists {
		return errors.New("chat not found")
	}

	chat.Messages = append(chat.Messages, message)
	return nil
}

// GetChat retrieves a conversation by its ChatUUID
func (m *MockChatHistoryStorage) GetChat(_ context.Context, uuid uuid.UUID) (*goai.ChatHistory, error) {
	if m.getChatErr != nil {
		return nil, m.getChatErr
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	chat, exists := m.chatLogs[uuid]
	if !exists {
		return nil, errors.New("chat not found")
	}

	return chat, nil
}

// ListChatHistories returns all stored conversations
func (m *MockChatHistoryStorage) ListChatHistories(_ context.Context) ([]goai.ChatHistory, error) {
	if m.listChatHistoriesErr != nil {
		return nil, m.listChatHistoriesErr
	}

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	chats := make([]goai.ChatHistory, 0, len(m.chatLogs))
	for _, chat := range m.chatLogs {
		chats = append(chats, *chat)
	}

	return chats, nil
}

// DeleteChat removes a conversation by its ChatUUID
func (m *MockChatHistoryStorage) DeleteChat(_ context.Context, uuid uuid.UUID) error {
	if m.deleteChatErr != nil {
		return m.deleteChatErr
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.chatLogs[uuid]; !exists {
		return errors.New("chat not found")
	}

	delete(m.chatLogs, uuid)
	return nil
}

// SetCreateChatError sets the error to be returned by CreateChat
func (m *MockChatHistoryStorage) SetCreateChatError(err error) {
	m.createChatErr = err
}

// SetAddMessageError sets the error to be returned by AddMessage
func (m *MockChatHistoryStorage) SetAddMessageError(err error) {
	m.addMessageErr = err
}

// SetGetChatError sets the error to be returned by GetChat
func (m *MockChatHistoryStorage) SetGetChatError(err error) {
	m.getChatErr = err
}

// SetListChatHistoriesError sets the error to be returned by ListChatHistories
func (m *MockChatHistoryStorage) SetListChatHistoriesError(err error) {
	m.listChatHistoriesErr = err
}

// SetDeleteChatError sets the error to be returned by DeleteChat
func (m *MockChatHistoryStorage) SetDeleteChatError(err error) {
	m.deleteChatErr = err
}

// ResetErrors resets all error simulation fields
func (m *MockChatHistoryStorage) ResetErrors() {
	m.createChatErr = nil
	m.addMessageErr = nil
	m.getChatErr = nil
	m.listChatHistoriesErr = nil
	m.deleteChatErr = nil
}

// Clear removes all stored chat histories
func (m *MockChatHistoryStorage) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.chatLogs = make(map[uuid.UUID]*goai.ChatHistory)
}
