package theme

import (
	"bytes"
	"github.com/fatih/color"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDisplayBanner(t *testing.T) {
	// Create a test config
	appCfg := &config.AppConfig{
		Name: "EchoY",
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Save the original color.Output
	originalOutput := color.Output

	// Set color.Output to our buffer
	color.Output = &buf

	// Restore the original color.Output when the test finishes
	defer func() {
		color.Output = originalOutput
	}()

	// Call DisplayBanner
	DisplayBanner(appCfg)

	// Get the captured output
	output := buf.String()

	// Log captured output for debugging
	t.Logf("Captured output:\n%s", output)

	// Check if output contains expected elements
	assert.Contains(t, output, "Welcome to EchoY")
	assert.Contains(t, output, "Your AI assistant for the CLI")
	assert.Contains(t, output, "╔═") // Start of top border
	assert.Contains(t, output, "╚═") // Start of bottom border
}
