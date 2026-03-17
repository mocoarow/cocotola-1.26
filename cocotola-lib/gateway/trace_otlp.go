package gateway

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

func initTracerExporterOTLPHTTP(ctx context.Context, traceConfig TraceConfig) (*otlptrace.Exporter, error) {
	if traceConfig.OTLP == nil {
		return nil, errors.New("otlp trace configuration is required")
	}

	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(traceConfig.OTLP.Endpoint),
	}
	if traceConfig.OTLP.Insecure {
		options = append(options, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP HTTP trace exporter: %w", err)
	}

	return exp, nil
}

func initTracerExporterOTLPgRPC(ctx context.Context, traceConfig TraceConfig) (*otlptrace.Exporter, error) {
	if traceConfig.OTLP == nil {
		return nil, errors.New("otlp trace configuration is required")
	}

	options := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(traceConfig.OTLP.Endpoint),
	}
	if traceConfig.OTLP.Insecure {
		options = append(options, otlptracegrpc.WithInsecure())
	}

	exp, err := otlptracegrpc.New(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP gRPC trace exporter: %w", err)
	}

	return exp, nil
}
