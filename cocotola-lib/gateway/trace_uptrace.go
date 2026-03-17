package gateway

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func initTracerExporterUptraceHTTP(ctx context.Context, traceConfig TraceConfig) (*otlptrace.Exporter, error) {
	if traceConfig.Uptrace == nil {
		return nil, errors.New("uptrace trace configuration is required")
	}

	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(traceConfig.Uptrace.Endpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"uptrace-dsn": traceConfig.Uptrace.DSN,
		}),
		otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
	)
	if err != nil {
		return nil, fmt.Errorf("create Uptrace HTTP trace exporter: %w", err)
	}

	return exp, nil
}
