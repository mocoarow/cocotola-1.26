package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

// serverlessAuthTransport sends the Google Cloud ID token in
// X-Serverless-Authorization so Authorization stays free for app-level tokens.
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
		return nil, errors.New("obtain ID token: token is empty")
	}

	req = req.Clone(req.Context())
	req.Header.Set("X-Serverless-Authorization", "Bearer "+token.AccessToken)

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}

	return resp, nil
}

// NewHTTPClient creates an otelhttp-instrumented client for service-to-service
// calls. Outside "local"/"test" it attaches a Google Cloud ID token.
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
		Transport:     otelhttp.NewTransport(sat),
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       timeout,
	}, nil
}
