package gateway

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

const (
	googleTraceEndpoint = "telemetry.googleapis.com:443"
	googleTraceScope    = "https://www.googleapis.com/auth/cloud-platform"
)

func initTracerExporterGoogle(ctx context.Context, traceConfig TraceConfig) (*otlptrace.Exporter, error) {
	if traceConfig.Google == nil || traceConfig.Google.ProjectID == "" {
		return nil, ErrGoogleTraceProjectIDRequired
	}

	creds, err := oauth.NewApplicationDefault(ctx, googleTraceScope)
	if err != nil {
		return nil, fmt.Errorf("find application default credentials: %w", err)
	}

	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(googleTraceEndpoint),
		// Set TLS explicitly so OTEL_EXPORTER_OTLP_*INSECURE cannot downgrade the connection.
		otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
		otlptracegrpc.WithDialOption(grpc.WithPerRPCCredentials(creds)),
	)
	if err != nil {
		return nil, fmt.Errorf("create Cloud Trace OTLP exporter: %w", err)
	}

	return exp, nil
}
