package telemetry

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/shaharia-lab/echoy/internal/config"
	"github.com/shaharia-lab/telemetry-collector"
	"runtime"
	"runtime/debug"
)

const (
	telemetryEndpoint = "https://telemetry-pub.shaharialab.com/telemetry/event"
)

// SendTelemetryEvent sends a telemetry event to the specified endpoint
func SendTelemetryEvent(ctx context.Context, appCfg *config.AppConfig, eventName string, severityText telemetry.Severity, message string, attributes map[string]interface{}) {
	collector := telemetry.NewCollector(
		telemetryEndpoint,
		fmt.Sprintf("%s-cli", appCfg.Name),
	)
	defer collector.Close()

	telemetryEvent := telemetry.Event{
		Name:         eventName,
		TraceID:      uuid.New().String(),
		SpanID:       uuid.New().String(),
		SeverityText: severityText,
		Body:         message,
		Attributes: map[string]interface{}{
			"cli_version.code":   appCfg.Version.Version,
			"cli_version.commit": appCfg.Version.Commit,
			"cli_version.date":   appCfg.Version.Date,
			"os.name":            runtime.GOOS,
			"os.arch":            runtime.GOARCH,
			"go.runtime_version": runtime.Version(),
		},
		Resource: map[string]interface{}{
			"service.name":    appCfg.Name,
			"service.version": appCfg.Version.Version,
		},
	}

	buildInfo, _ := debug.ReadBuildInfo()
	for _, setting := range buildInfo.Settings {
		key := fmt.Sprintf("build_settings.%s", setting.Key)
		telemetryEvent.Attributes[key] = setting.Value
	}

	for k, v := range attributes {
		if _, exists := telemetryEvent.Attributes[k]; exists {
			continue
		}
		telemetryEvent.Attributes[k] = fmt.Sprintf("%v", v)
	}

	collector.SendAsync(ctx, &telemetryEvent)
}
