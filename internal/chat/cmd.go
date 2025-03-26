package chat

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/goai"
	"github.com/spf13/cobra"
)

// NewChatCmd creates a new chat command
func NewChatCmd(c *cli.Container) *cobra.Command {
	cmd := &cobra.Command{
		Version: c.Config.Version.VersionText(),
		Use:     "chat",
		Short:   "Start an interactive chat session",
		Long:    `Begin an interactive chat session with Echoy. Each session is uniquely identified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startChatSession(c.ThemeMgr.GetCurrentTheme())
		},
	}

	return cmd
}

// startChatSession begins an interactive chat session
func startChatSession(theme theme.Theme) error {
	cfgManager := &initPkg.DefaultConfigManager{}
	cfg, err := cfgManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	llmService, err := NewLLMService(cfg.LLM)
	if err != nil {
		return fmt.Errorf("error initializing LLM service: %w", err)
	}

	historyService := goai.NewInMemoryChatHistoryStorage()
	chatService := NewChatService(llmService, historyService)
	session, err := NewChatSession(&cfg, theme, chatService, historyService)
	if err != nil {
		return fmt.Errorf("error creating chat session: %w", err)
	}

	return session.Start(context.Background())
}
