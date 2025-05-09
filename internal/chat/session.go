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

// Session represents an interactive chat session
type Session struct {
	config                *config.Config
	theme                 theme.Theme
	chatService           Service
	chatHistoryStorage    goai.ChatHistoryStorage
	sessionID             uuid.UUID
	reader                *bufio.Reader
	chatHistoryService    HistoryService
	thinkingAnimationFunc func(theme theme.Theme, thinking chan bool)
}

// NewChatSession creates and configures a new chat session
func NewChatSession(config *config.Config, theme theme.Theme, chatService Service, chatHistoryService HistoryService) (*Session, error) {
	ctx := context.Background()

	sessionID, err := chatHistoryService.CreateChat(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating chat session: %w", err)
	}

	return &Session{
		config:                config,
		theme:                 theme,
		chatService:           chatService,
		chatHistoryService:    chatHistoryService,
		sessionID:             sessionID.UUID,
		reader:                bufio.NewReader(os.Stdin),
		thinkingAnimationFunc: showThinkingAnimation,
	}, nil
}

// Start begins the interactive chat session
func (s *Session) Start(ctx context.Context) error {
	s.showWelcomeMessage()

	for {
		input, err := s.readUserInput()
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		if strings.ToLower(input) == "exit" {
			s.theme.Info().Println("Ending chat session. Goodbye")
			return nil
		}

		if strings.ToLower(input) == "clear" {
			fmt.Print("\033[H\033[2J")
			continue
		}

		if s.config.LLM.Streaming {
			if err := s.processMessageStreaming(ctx, input); err != nil {
				return err
			}

			continue
		}

		if err := s.processMessage(ctx, input); err != nil {
			return err
		}
	}
}

func (s *Session) showWelcomeMessage() {
	s.theme.Info().Println("\n🗨️ Chat session started.")
	s.theme.Subtle().Println("Session ID: ", s.sessionID)
	s.theme.Secondary().Println("Type your message and press Enter. For multi-line input, continue typing.")
	s.theme.Secondary().Println("Press Enter twice (empty line) to submit your message.")
	s.theme.Secondary().Println("Type 'exit' to end the session.")
}

func (s *Session) readUserInput() (string, error) {
	s.theme.Primary().Print(fmt.Sprintf("%s > ", s.config.User.Name))

	var builder strings.Builder
	var lines []string
	isSubmitting := false

	for !isSubmitting {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		trimmedLine := strings.TrimSpace(line)
		lines = append(lines, trimmedLine)

		if trimmedLine == "" && len(lines) > 1 {
			isSubmitting = true
		}
	}

	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}

		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(line)
	}

	return builder.String(), nil
}

func (s *Session) processMessage(ctx context.Context, input string) error {
	thinking := make(chan bool, 1)
	defer close(thinking)

	go s.thinkingAnimationFunc(s.theme, thinking)

	response, err := s.chatService.Chat(ctx, s.sessionID, input)
	if err != nil {
		thinking <- true
		return fmt.Errorf("error processing chat input: %w", err)
	}

	thinking <- true
	fmt.Print("\r                \r")

	s.theme.Secondary().Print("AI > ")
	s.theme.Subtle().Printf("%s\n", response.Answer)

	return nil
}

func (s *Session) processMessageStreaming(ctx context.Context, input string) error {
	thinking := make(chan bool, 1)
	defer close(thinking)

	go s.thinkingAnimationFunc(s.theme, thinking)

	streamChan, err := s.chatService.ChatStreaming(ctx, s.sessionID, input)
	if err != nil {
		thinking <- true
		return fmt.Errorf("error processing chat input: %w", err)
	}

	firstToken := true
	s.theme.Secondary().Print("AI > ")

	for streamResp := range streamChan {
		if firstToken {
			thinking <- true
			fmt.Print("\r                \r")
			s.theme.Secondary().Print("AI > ")
			firstToken = false
		}

		if streamResp.Error != nil {
			return fmt.Errorf("error in streaming response: %w", streamResp.Error)
		}

		s.theme.Subtle().Print(streamResp.Text)
	}

	fmt.Println()
	return nil
}

func showThinkingAnimation(theme theme.Theme, thinking chan bool) {
	go func() {
		dots := []string{".  ", ".. ", "..."}
		i := 0
		for {
			select {
			case <-thinking:
				return
			default:
				theme.Warning().Printf("\rThinking%s", dots[i%3])
				i++
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
}
