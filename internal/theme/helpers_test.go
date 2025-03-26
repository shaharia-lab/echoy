package theme

import (
	"github.com/shaharia-lab/echoy/internal/config"
	"testing"
)

func TestDisplayBanner(t *testing.T) {
	// Create a test config
	appCfg := &config.AppConfig{
		Name: "EchoY",
	}

	// Call DisplayBanner and visually inspect the output
	// This is a simpler test that just verifies it runs without errors
	DisplayBanner(appCfg)
}
