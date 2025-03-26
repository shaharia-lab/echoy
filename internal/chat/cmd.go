package chat

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/config"
	initPkg "github.com/shaharia-lab/echoy/internal/init"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/goai"
	"github.com/spf13/cobra"
)

// NewChatCmd creates a new chat command
func NewChatCmd(appCfg *config.AppConfig, log *logger.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "chat",
		Short:   "Start an interactive chat session",
		Long:    `Begin an interactive chat session with Echoy. Each session is uniquely identified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return startChatSession()
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

	llmService, err := NewLLMService(cfg.LLM)
	if err != nil {
		return fmt.Errorf("error initializing LLM service: %w", err)
	}

	historyService := goai.NewInMemoryChatHistoryStorage()
	chatService := NewChatService(llmService, historyService)
	session, err := NewChatSession(&cfg, cliTheme, chatService, historyService)
	if err != nil {
		return fmt.Errorf("error creating chat session: %w", err)
	}

	return session.Start(context.Background())
}
