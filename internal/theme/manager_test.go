package theme_test

import (
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/echoy/internal/theme/mocks"
	"testing"
)

func TestManager_DisplayBanner(t *testing.T) {
	tests := []struct {
		name       string
		title      string
		width      int
		subtitle   []string
		setupMocks func(mockTheme *mocks.MockTheme, primaryMock *mocks.MockStylePrinter, secondaryMock *mocks.MockStylePrinter)
	}{
		{
			name:     "Basic title without subtitle",
			title:    "My App",
			width:    20,
			subtitle: []string{},
			setupMocks: func(mockTheme *mocks.MockTheme, primaryMock *mocks.MockStylePrinter, secondaryMock *mocks.MockStylePrinter) {
				// Setup theme mock to return style mocks
				mockTheme.On("Primary").Return(primaryMock)
				mockTheme.On("Secondary").Return(secondaryMock)

				// Top border
				primaryMock.On("Println", "╔══════════════════╗").Once()

				// Title line
				primaryMock.On("Println", "║      My App      ║").Once()

				// Bottom border
				primaryMock.On("Println", "╚══════════════════╝").Once()
			},
		},
		{
			name:     "Title with subtitle",
			title:    "My App",
			width:    20,
			subtitle: []string{"Version 1.0"},
			setupMocks: func(mockTheme *mocks.MockTheme, primaryMock *mocks.MockStylePrinter, secondaryMock *mocks.MockStylePrinter) {
				// Setup theme mock to return style mocks
				mockTheme.On("Primary").Return(primaryMock)
				mockTheme.On("Secondary").Return(secondaryMock)

				// Top border
				primaryMock.On("Println", "╔══════════════════╗").Once()

				// Title line
				primaryMock.On("Println", "║      My App      ║").Once()

				// Add separator line expectation
				primaryMock.On("Println", "║──────────────────║").Once()

				// Subtitle line - adjust spacing to match implementation
				secondaryMock.On("Println", "║   Version 1.0    ║").Once()

				// Bottom border
				primaryMock.On("Println", "╚══════════════════╝").Once()
			},
		},
		{
			name:     "Title with multiple subtitles",
			title:    "My App",
			width:    20,
			subtitle: []string{"Version 1.0", "By Developer"},
			setupMocks: func(mockTheme *mocks.MockTheme, primaryMock *mocks.MockStylePrinter, secondaryMock *mocks.MockStylePrinter) {
				// Setup theme mock to return style mocks
				mockTheme.On("Primary").Return(primaryMock)
				mockTheme.On("Secondary").Return(secondaryMock)

				// Top border
				primaryMock.On("Println", "╔══════════════════╗").Once()

				// Title line
				primaryMock.On("Println", "║      My App      ║").Once()

				// Add separator line expectation
				primaryMock.On("Println", "║──────────────────║").Once()

				// Subtitle lines - adjust spacing to match implementation
				secondaryMock.On("Println", "║   Version 1.0    ║").Once()
				secondaryMock.On("Println", "║   By Developer   ║").Once()

				// Bottom border
				primaryMock.On("Println", "╚══════════════════╝").Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockTheme := new(mocks.MockTheme)
			primaryMock := new(mocks.MockStylePrinter)
			secondaryMock := new(mocks.MockStylePrinter)

			// Setup mock expectations
			tt.setupMocks(mockTheme, primaryMock, secondaryMock)

			// Create manager with mocks
			manager := theme.NewManager(mockTheme, &config.AppConfig{}, nil)

			// Call the method under test
			manager.DisplayBanner(tt.title, tt.width, tt.subtitle...)

			// Verify all expectations were met
			mockTheme.AssertExpectations(t)
			primaryMock.AssertExpectations(t)
			secondaryMock.AssertExpectations(t)
		})
	}
}
