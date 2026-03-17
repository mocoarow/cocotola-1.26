package gateway

import (
	"context"
	"errors"
	"fmt"

	gcpexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
)

func initTracerExporterGoogle(_ context.Context, traceConfig TraceConfig) (*gcpexporter.Exporter, error) {
	if traceConfig.Google == nil {
		return nil, errors.New("google trace configuration is required")
	}

	exp, err := gcpexporter.New(gcpexporter.WithProjectID(traceConfig.Google.ProjectID))
	if err != nil {
		return nil, fmt.Errorf("create Google Cloud trace exporter: %w", err)
	}

	return exp, nil
}
