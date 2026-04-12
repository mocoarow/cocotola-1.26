package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// serverlessAuthTransport is an http.RoundTripper that attaches
// a Google Cloud ID token via the X-Serverless-Authorization header
// instead of the standard Authorization header.
// This allows the caller to use the Authorization header for
// application-level tokens (e.g. user access tokens).
type serverlessAuthTransport struct {
	base     http.RoundTripper
	tokenSrc oauth2.TokenSource
}

// NewServerlessAuthTransport creates a new serverlessAuthTransport.
func NewServerlessAuthTransport(base http.RoundTripper, tokenSrc oauth2.TokenSource) http.RoundTripper {
	return &serverlessAuthTransport{base: base, tokenSrc: tokenSrc}
}

func (t *serverlessAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.tokenSrc.Token()
	if err != nil {
		return nil, fmt.Errorf("obtain ID token: %w", err)
	}

	// idtoken.NewTokenSource stores the ID token JWT in oauth2.Token.AccessToken.
	// This is a Google library convention, not a standard OAuth2 pattern.
	if token.AccessToken == "" {
		return nil, fmt.Errorf("obtain ID token: token is empty")
	}

	req = req.Clone(req.Context())
	req.Header.Set("X-Serverless-Authorization", "Bearer "+token.AccessToken)

	return t.base.RoundTrip(req)
}

// NewHTTPClient creates an *http.Client suitable for service-to-service calls.
// The returned client wraps its transport with otelhttp so that OpenTelemetry
// trace context is propagated to downstream services.
// When appEnv is "local" or "test", it returns a plain client.
// Otherwise, it returns a client that attaches a Google Cloud ID token
// via the X-Serverless-Authorization header, leaving the Authorization header
// available for application-level tokens.
func NewHTTPClient(ctx context.Context, appEnv string, audience string, timeout time.Duration) (*http.Client, error) {
	if appEnv == "local" || appEnv == "test" {
		return &http.Client{
			Transport:     otelhttp.NewTransport(http.DefaultTransport),
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       timeout,
		}, nil
	}

	tokenSrc, err := idtoken.NewTokenSource(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("create ID token source: %w", err)
	}

	sat := NewServerlessAuthTransport(http.DefaultTransport, oauth2.ReuseTokenSource(nil, tokenSrc))

	return &http.Client{
		Transport: otelhttp.NewTransport(sat),
		Timeout:   timeout,
	}, nil
}
