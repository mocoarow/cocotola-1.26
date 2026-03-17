package gateway

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
)

func initLogExporterUptraceHTTP(ctx context.Context, logConfig LogConfig) (*otlploghttp.Exporter, error) {
	if logConfig.Uptrace == nil {
		return nil, errors.New("uptrace log configuration is required")
	}

	exp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(logConfig.Uptrace.Endpoint),
		otlploghttp.WithHeaders(map[string]string{
			"uptrace-dsn": logConfig.Uptrace.DSN,
		}),
		otlploghttp.WithCompression(otlploghttp.GzipCompression),
	)
	if err != nil {
		return nil, fmt.Errorf("create Uptrace HTTP log exporter: %w", err)
	}

	return exp, nil
}
