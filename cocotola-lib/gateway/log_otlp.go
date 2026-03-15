package gateway

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
)

func initLogExporterOTLPHTTP(ctx context.Context, logConfig *LogConfig) (*otlploghttp.Exporter, error) {
	if logConfig.OTLP == nil {
		return nil, errors.New("otlp log configuration is required")
	}

	options := []otlploghttp.Option{
		otlploghttp.WithEndpoint(logConfig.OTLP.Endpoint),
	}
	if logConfig.OTLP.Insecure {
		options = append(options, otlploghttp.WithInsecure())
	}

	exp, err := otlploghttp.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP log exporter: %w", err)
	}

	return exp, nil
}
