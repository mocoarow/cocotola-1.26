package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/api/idtoken"
)

// NewHTTPClient creates an *http.Client suitable for service-to-service calls.
// The returned client wraps its transport with otelhttp so that OpenTelemetry
// trace context is propagated to downstream services.
// When appEnv is "local" or "test", it returns a plain client.
// Otherwise, it returns a client that automatically attaches a Google Cloud ID token
// with the given audience (the target Cloud Run service URL).
func NewHTTPClient(ctx context.Context, appEnv string, audience string, timeout time.Duration) (*http.Client, error) {
	if appEnv == "local" || appEnv == "test" {
		return &http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport),
			Timeout:   timeout,
		}, nil
	}

	client, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("create ID token HTTP client: %w", err)
	}

	client.Transport = otelhttp.NewTransport(client.Transport)
	client.Timeout = timeout

	return client, nil
}
