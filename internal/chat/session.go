package chat

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/uuid"
	"os"
	"strings"
	"time"

	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/goai"
)

// ChatSession represents an interactive chat session
type ChatSession struct {
	config             *config.Config
	theme              theme.Theme
	chatService        *Service
	chatHistoryStorage goai.ChatHistoryStorage
	sessionID          uuid.UUID
	reader             *bufio.Reader
}

// NewChatSession creates and configures a new chat session
func NewChatSession(
	config *config.Config,
	theme theme.Theme,
	chatService *Service,
	chatHistoryStorage goai.ChatHistoryStorage,
) (*ChatSession, error) {
	ctx := context.Background()

	sessionID, err := chatHistoryStorage.CreateChat(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating chat session: %w", err)
	}

	return &ChatSession{
		config:             config,
		theme:              theme,
		chatService:        chatService,
		chatHistoryStorage: chatHistoryStorage,
		sessionID:          sessionID.UUID,
		reader:             bufio.NewReader(os.Stdin),
	}, nil
}

// Start begins the interactive chat session
func (s *ChatSession) Start(ctx context.Context) error {
	s.showWelcomeMessage()

	for {
		input, err := s.readUserInput()
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		// Handle special commands
		if strings.ToLower(input) == "exit" {
			s.theme.Info().Println("Ending chat session. Goodbye")
			return nil
		}

		if strings.ToLower(input) == "clear" {
			fmt.Print("\033[H\033[2J")
			continue
		}

		if err := s.processMessage(ctx, input); err != nil {
			return err
		}
	}
}

func (s *ChatSession) showWelcomeMessage() {
	s.theme.Info().Println("\n🗨️ Chat session started.")
	s.theme.Subtle().Println("Session ID: ", s.sessionID)
	s.theme.Secondary().Println("Type your message and press Enter. Type 'exit' to end the session.")
}

func (s *ChatSession) readUserInput() (string, error) {
	s.theme.Primary().Print(fmt.Sprintf("%s > ", s.config.User.Name))
	input, err := s.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(input), nil
}

func (s *ChatSession) processMessage(ctx context.Context, input string) error {
	thinking := make(chan bool)
	s.showThinkingAnimation(thinking)

	response, err := s.chatService.Chat(ctx, s.sessionID, input)
	if err != nil {
		return fmt.Errorf("error processing chat input: %w", err)
	}

	thinking <- true
	fmt.Print("\r                \r")

	s.theme.Secondary().Print("AI > ")
	s.theme.Subtle().Printf("%s\n", response.Text)

	return nil
}

func (s *ChatSession) showThinkingAnimation(thinking chan bool) {
	go func() {
		dots := []string{".  ", ".. ", "..."}
		i := 0
		for {
			select {
			case <-thinking:
				return
			default:
				s.theme.Warning().Printf("\rThinking%s", dots[i%3])
				i++
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
}
