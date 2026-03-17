package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/processors/baggagecopy"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const (
	traceMaxQueueSize      = 10_000
	traceExportTimeoutSec  = 10
	samplingPercentageFull = 100
	samplingPercentageDiv  = 100.0
	traceShutdownTimeout   = 5 * time.Second
)

// OTLPTraceConfig holds OTLP trace exporter settings.
type OTLPTraceConfig struct {
	Endpoint string `yaml:"endpoint" validate:"required"`
	Insecure bool   `yaml:"insecure"`
}

// GoogleTraceConfig holds Google Cloud trace exporter settings.
type GoogleTraceConfig struct {
	ProjectID string `yaml:"projectId" validate:"required"`
}

// UptraceTraceConfig holds Uptrace trace exporter settings.
type UptraceTraceConfig struct {
	Endpoint string `yaml:"endpoint" validate:"required"`
	DSN      string `yaml:"dsn" validate:"required"`
}

// TraceConfig holds the exporter type, sampling percentage, and provider-specific settings.
type TraceConfig struct {
	Exporter           string              `yaml:"exporter" validate:"required,oneof=otlphttp otlpgrpc uptracehttp google stdout none"`
	SamplingPercentage int                 `yaml:"samplingPercentage" validate:"gte=0,lte=100"`
	OTLP               *OTLPTraceConfig    `yaml:"otlp"`
	Google             *GoogleTraceConfig  `yaml:"google"`
	Uptrace            *UptraceTraceConfig `yaml:"uptrace"`
}

func initTracerExporter(ctx context.Context, traceConfig TraceConfig) (sdktrace.SpanExporter, error) { //nolint:ireturn // returns interface required by OpenTelemetry SDK
	switch traceConfig.Exporter {
	case "google":
		return initTracerExporterGoogle(ctx, traceConfig)
	case "otlphttp":
		return initTracerExporterOTLPHTTP(ctx, traceConfig)
	case "otlpgrpc":
		return initTracerExporterOTLPgRPC(ctx, traceConfig)
	case "none":
		return initTracerExporterNone(ctx, traceConfig)
	case "stdout":
		return initTracerExporterStdout(ctx, traceConfig)
	case "uptracehttp":
		return initTracerExporterUptraceHTTP(ctx, traceConfig)
	default:
		return nil, fmt.Errorf("invalid trace exporter: %s", traceConfig.Exporter)
	}
}

func initTraceSampler(samplingPercentage int) sdktrace.Sampler { //nolint:ireturn // returns interface required by OpenTelemetry SDK
	if samplingPercentage >= samplingPercentageFull {
		return sdktrace.AlwaysSample()
	}
	if samplingPercentage <= 0 {
		return sdktrace.NeverSample()
	}
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(float64(samplingPercentage) / samplingPercentageDiv))
}

// InitTracerProvider creates an OpenTelemetry tracer provider and sets it as the global default.
func InitTracerProvider(ctx context.Context, traceConfig TraceConfig, appName string) (func(), error) {
	exp, err := initTracerExporter(ctx, traceConfig)
	if err != nil {
		return nil, fmt.Errorf("init tracer exporter: %w", err)
	}

	sampler := initTraceSampler(traceConfig.SamplingPercentage)

	bp := sdktrace.NewBatchSpanProcessor(exp,
		sdktrace.WithMaxQueueSize(traceMaxQueueSize),
		sdktrace.WithMaxExportBatchSize(traceMaxQueueSize),
		sdktrace.WithExportTimeout(traceExportTimeoutSec*time.Second),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bp),
		sdktrace.WithSpanProcessor(baggagecopy.NewSpanProcessor(baggagecopy.AllowAllMembers)),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(appName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		shutdownBaseCtx := context.Background()
		if ctx != nil {
			shutdownBaseCtx = context.WithoutCancel(ctx)
		}
		shutdownCtx, cancel := context.WithTimeout(shutdownBaseCtx, traceShutdownTimeout)
		defer cancel()

		if err := tp.Shutdown(shutdownCtx); err != nil {
			slog.Default().Error("shutdown tracer provider", slog.Any("error", err))
		}
	}, nil
}
