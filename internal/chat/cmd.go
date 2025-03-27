package chat

import (
	"context"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/spf13/cobra"
)

// NewChatCmd creates a new chat command
func NewChatCmd(appCfg *config.AppConfig, chatSession *Session) *cobra.Command {
	cmd := &cobra.Command{
		Version: appCfg.Version.VersionText(),
		Use:     "chat",
		Short:   "Start an interactive chat session",
		Long:    `Begin an interactive chat session with Echoy. Each session is uniquely identified.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return chatSession.Start(context.Background())
		},
	}

	return cmd
}
