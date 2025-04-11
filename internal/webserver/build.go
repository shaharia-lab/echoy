package webserver

import (
	"fmt"
	"github.com/shaharia-lab/echoy/internal/chat"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/echoy/internal/llm"
	"github.com/shaharia-lab/echoy/internal/logger"
	"github.com/shaharia-lab/echoy/internal/theme"
	"github.com/shaharia-lab/echoy/internal/tools"
	"github.com/shaharia-lab/echoy/internal/webui"
	"github.com/shaharia-lab/goai"
	"github.com/shaharia-lab/goai/mcp"
	mcpTools "github.com/shaharia-lab/mcp-tools"
	"net/http"
)

// BuildWebserver initializes the web server with the provided configuration and dependencies
func BuildWebserver(config config.Config, themeManager *theme.Manager, webUIStaticDirectory string, logDirectory string) (*WebServer, error) {
	serverLogger, err := logger.NewZapLogger(logger.Config{
		LogLevel:    logger.DebugLevel,
		LogFilePath: fmt.Sprintf("%s/webserver.log", logDirectory),

		MaxSizeMB:  50,
		MaxAgeDays: 14,
		MaxBackups: 5,
	})
	if err != nil {
		themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to initialize webserver logger: %v", err))
		return nil, fmt.Errorf("failed to initialize webserver logger: %w", err)
	}

	ts := []mcp.Tool{
		mcpTools.GetWeather,
	}

	llmService, err := llm.NewLLMService(config.LLM)
	if err != nil {
		serverLogger.Errorf("Failed to create LLM service: %v", err)
		themeManager.GetCurrentTheme().Error().Println(fmt.Sprintf("Failed to create LLM service: %v", err))
		return nil, err
	}

	historyService := goai.NewInMemoryChatHistoryStorage()

	chatService := chat.NewChatService(llmService, historyService)
	chatHandler := chat.NewChatHandler(chatService)
	webUIDownloaderHttpClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	return NewWebServer(
		"10222",
		webUIStaticDirectory,
		tools.NewProvider(ts),
		llm.NewLLMHandler(llm.GetSupportedLLMProviders()),
		chatHandler,
		webui.NewFrontendGitHubReleaseDownloader(webUIStaticDirectory, webUIDownloaderHttpClient, serverLogger),
	), nil
}
