package chat

import (
	"context"
	"fmt"
	"github.com/shaharia-lab/echoy/internal/cli"
	"github.com/shaharia-lab/echoy/internal/llm"
	telemetryEvent "github.com/shaharia-lab/echoy/internal/telemetry"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/telemetry-collector"
	"github.com/spf13/cobra"
)

// NewChatCmd creates a new chat command
func NewChatCmd(container *cli.Container) *cobra.Command {
	cmd := &cobra.Command{
		Version: container.Config.Version.VersionText(),
		Use:     "chat",
		Short:   "Start an interactive chat session",
		Long:    `Begin an interactive chat session with Echoy. Each session is uniquely identified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			llmService, err := llm.NewLLMService(container.ConfigFromFile.LLM)
			if err != nil {
				container.Logger.Errorf(fmt.Sprintf("error initializing LLM service: %v", err))
				return fmt.Errorf("error initializing LLM service: %w", err)
			}

			chatHistoryService := goai.NewInMemoryChatHistoryStorage()
			chatService := NewChatService(llmService, chatHistoryService)
			chatSession, err := NewChatSession(&container.ConfigFromFile, container.ThemeMgr.GetCurrentTheme(), chatService, chatHistoryService)
			if err != nil {
				container.Logger.Errorf(fmt.Sprintf("error creating chat session: %v", err))
				return fmt.Errorf("error creating chat session: %w", err)
			}

			ctx := context.Background()
			telemetryEvent.SendTelemetryEvent(
				ctx,
				container.Config,
				"cmd.chat",
				telemetry.SeverityInfo, "Starting chat session",
				nil,
			)

			return chatSession.Start(ctx)
		},
	}

	return cmd
}
