package cmd

import (
	"bufio"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/config"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/goai"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

// NewChatCmd creates a new chat command
func NewChatCmd(appCfg *config.AppConfig) *cobra.Command {
	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "chat",
		Short:   "Start an interactive chat session",
		Long:    `Begin an interactive chat session with Echoy. Each session is uniquely identified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			startChatSession()
			return nil
		},
	}

	return cmd
}

// startChatSession begins an interactive chat session
func startChatSession() {
	cfgManager := &initPkg.DefaultConfigManager{}
	cfg, err := cfgManager.LoadConfig()
	if err != nil {
		color.Red("Error loading configuration: %v", err)
		return
	}
	ctx := context.Background()

	chatHistoryStorage := goai.NewInMemoryChatHistoryStorage()
	sessionID, err := chatHistoryStorage.CreateChat(ctx)
	if err != nil {
		color.Red("Error creating chat session: %v", err)
		return
	}

	color.New(color.FgHiCyan, color.Bold).Println("\n🗨️  New Chat Session")
	color.New(color.FgHiWhite).Printf("Session ID: %s\n\n", sessionID.UUID)
	color.Yellow("Type your message and press Enter. Type 'exit' to end the session.")

	// Set up colors for different elements
	promptColor := color.New(color.FgHiGreen, color.Bold)
	systemMsgColor := color.New(color.FgCyan)
	headerColor := color.New(color.FgHiMagenta, color.Bold)

	// Display a nicer header
	headerColor.Println("\n✨ ========================================= ✨")
	headerColor.Println("       P L E A S E   A I   C H A T          ")
	headerColor.Println("✨ ========================================= ✨")

	systemMsgColor.Printf("\nSession ID: %s\n\n", sessionID.UUID)
	systemMsgColor.Println("Type your message and press Enter. Type 'exit' to end the session.")
	systemMsgColor.Println("Type 'clear' to clear the screen.\n")

	reader := bufio.NewReader(os.Stdin)

	for {
		// Show a better prompt with indicator of who's speaking
		promptColor.Print("You > ")
		input, err := reader.ReadString('\n')

		if err != nil {
			color.Red("Error reading input: %v", err)
			return
		}

		// Trim whitespace and convert to lowercase for command checking
		input = strings.TrimSpace(input)
		// Add commands like clear screen
		if strings.ToLower(input) == "clear" {
			// Clear the screen - this works on most terminals
			fmt.Print("\033[H\033[2J")
			continue
		}

		if strings.ToLower(input) == "exit" {
			color.Yellow("Ending chat session %s. Goodbye!", sessionID.UUID)
			return
		}

		// Show thinking indicator
		thinking := make(chan bool)
		go func() {
			dots := []string{".  ", ".. ", "..."}
			i := 0
			for {
				select {
				case <-thinking:
					return
				default:
					systemMsgColor.Printf("\rThinking%s", dots[i%3])
					i++
					time.Sleep(300 * time.Millisecond)
				}
			}
		}()

		chatService, err := chat.NewChatService(cfg.LLM, chatHistoryStorage)
		if err != nil {
			color.Red("Error initializing chat service: %v", err)
			return
		}

		response, err := chatService.Chat(ctx, sessionID.UUID, input)
		if err != nil {
			color.Red("Error processing chat input: %v", err)
			return
		}
		thinking <- true
		fmt.Print("\r                \r")

		// Display AI response with AI indicator and better formatting
		color.New(color.FgHiCyan, color.Bold).Print("AI > ")
		color.New(color.FgHiBlack).Printf("%s\n", response.Text)

		// Add a separator between exchanges
		color.New(color.Faint).Println("")

	}
}
