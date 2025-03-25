package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/config"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/goai"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

// NewChatCmd creates a new chat command
func NewChatCmd(appCfg *config.AppConfig, log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "chat",
		Short:   "Start an interactive chat session",
		Long:    `Begin an interactive chat session with Echoy. Each session is uniquely identified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := startChatSession()
			if err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

// startChatSession begins an interactive chat session
func startChatSession() error {
	cliTheme := cli.GetTheme()

	cfgManager := &initPkg.DefaultConfigManager{}
	cfg, err := cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}
	ctx := context.Background()

	chatHistoryStorage := goai.NewInMemoryChatHistoryStorage()
	sessionID, err := chatHistoryStorage.CreateChat(ctx)
	if err != nil {
		return fmt.Errorf("error creating chat session: %w", err)
	}

	cliTheme.Info().Println("\n🗨️ Chat session started.")
	cliTheme.Subtle().Println("Session ID: ", sessionID.UUID)
	cliTheme.Secondary().Println("Type your message and press Enter. Type 'exit' to end the session.")

	reader := bufio.NewReader(os.Stdin)

	for {
		cliTheme.Primary().Print(fmt.Sprintf("%s > ", cfg.User.Name))
		input, err := reader.ReadString('\n')

		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)
		if strings.ToLower(input) == "clear" {
			fmt.Print("\033[H\033[2J")
			continue
		}

		if strings.ToLower(input) == "exit" {
			cliTheme.Info().Println("Ending chat session. Goodbye")
			return nil
		}

		thinking := make(chan bool)
		loadThinking(thinking, cliTheme)

		chatService, err := chat.NewChatService(cfg.LLM, chatHistoryStorage)
		if err != nil {
			return fmt.Errorf("error initializing chat service: %w", err)
		}

		response, err := chatService.Chat(ctx, sessionID.UUID, input)
		if err != nil {
			return fmt.Errorf("error processing chat input: %w", err)
		}
		thinking <- true
		fmt.Print("\r                \r")

		cliTheme.Secondary().Print("AI > ")
		cliTheme.Subtle().Printf("%s\n", response.Text)
	}
}

// loadThinking displays a thinking indicator while the AI is processing
func loadThinking(thinking chan bool, cliTheme theme.Theme) {
	go func() {
		dots := []string{".  ", ".. ", "..."}
		i := 0
		for {
			select {
			case <-thinking:
				return
			default:
				cliTheme.Warning().Printf("\rThinking%s", dots[i%3])
				i++
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
}
