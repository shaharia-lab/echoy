package logger

import "github.com/stretchr/testify/mock"

// MockLogger implements a test-friendly logger
type MockLogger struct {
	mock.Mock
}

// Debug mocks the Debug method
func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

// Info mocks the Info method
func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

// Warn mocks the Warn method
func (m *MockLogger) Warn(msg string) {
	m.Called(msg)
}

// Error mocks the Error method
func (m *MockLogger) Error(msg string) {
	m.Called(msg)
}

// Fatal mocks the Fatal method
func (m *MockLogger) Fatal(msg string) {
	m.Called(msg)
}
