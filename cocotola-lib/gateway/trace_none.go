package gateway

import (
	"context"
	"fmt"
	"io"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

func initTracerExporterNone(_ context.Context, _ TraceConfig) (*stdouttrace.Exporter, error) {
	exp, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(io.Discard),
	)
	if err != nil {
		return nil, fmt.Errorf("create no-op trace exporter: %w", err)
	}

	return exp, nil
}
