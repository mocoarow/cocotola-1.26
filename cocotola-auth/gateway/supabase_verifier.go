package gateway

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	liblogging "github.com/mocoarow/cocotola-1.26/cocotola-lib/logging"
)

// allowedSupabaseAlgs is asymmetric-only: HS* and "none" must never be accepted.
func allowedSupabaseAlgs() []string {
	return []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}
}

// SupabaseVerifier verifies Supabase JWT tokens using JWKS.
type SupabaseVerifier struct {
	jwks          keyfunc.Keyfunc
	jwksAvailable *atomic.Bool
	cancel        context.CancelFunc
}

// NewSupabaseVerifier returns a SupabaseVerifier that fetches the JWKS from the given URL.
func NewSupabaseVerifier(ctx context.Context, jwksURL string) (*SupabaseVerifier, error) {
	ctx, cancel := context.WithCancel(ctx)
	logger := slog.Default().With(slog.String(liblogging.LoggerNameKey, domain.AppName+"-SupabaseVerifier"))
	jwksAvailable := &atomic.Bool{}

	jwks, err := keyfunc.NewDefaultOverrideCtx(ctx, []string{jwksURL}, newJWKSOverride(jwksAvailable, logger))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create JWKS keyfunc for %s: %w", jwksURL, err)
	}

	return &SupabaseVerifier{jwks: jwks, jwksAvailable: jwksAvailable, cancel: cancel}, nil
}

func newJWKSOverride(available *atomic.Bool, logger *slog.Logger) keyfunc.Override {
	return keyfunc.Override{
		Client: &http.Client{
			Transport:     &jwksTransport{base: http.DefaultTransport, available: available},
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		},
		HTTPTimeout:               0,
		NoErrorReturnFirstHTTPReq: nil,
		RateLimitWaitMax:          0,
		RefreshErrorHandlerFunc: func(u string) func(context.Context, error) {
			return func(ctx context.Context, err error) {
				if errors.Is(err, context.Canceled) {
					return
				}
				available.Store(false)
				logger.ErrorContext(ctx, "refresh JWKS", slog.String("url", u), slog.Any("error", err))
			}
		},
		RefreshInterval:   0,
		RefreshUnknownKID: nil,
		ValidationSkipAll: false,
	}
}

type jwksTransport struct {
	base      http.RoundTripper
	available *atomic.Bool
}

func (t *jwksTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS round trip: %w", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.available.Store(true)
	}
	return resp, nil
}

// Verify validates a Supabase JWT and returns its sub and email.
// A valid token without email_verified=true yields ErrSupabaseEmailNotVerified;
// callers MUST NOT trust it (spoofed email enables takeover via auto-linking).
func (v *SupabaseVerifier) Verify(ctx context.Context, tokenString string) (string, string, error) {
	token, err := jwt.Parse(
		tokenString,
		v.jwks.KeyfuncCtx(ctx),
		jwt.WithValidMethods(allowedSupabaseAlgs()),
	)
	if err != nil {
		if errors.Is(err, keyfunc.ErrKeyfunc) && !v.jwksAvailable.Load() {
			return "", "", fmt.Errorf("parse supabase token: %w: %w", err, domain.ErrInternal)
		}
		return "", "", fmt.Errorf("parse supabase token: %w: %w", err, domain.ErrUnauthenticated)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid supabase token claims: %w", domain.ErrUnauthenticated)
	}

	sub, _ := claims.GetSubject()
	if sub == "" {
		return "", "", fmt.Errorf("supabase token missing sub claim: %w", domain.ErrUnauthenticated)
	}

	email, _ := claims["email"].(string)
	if email == "" {
		return "", "", fmt.Errorf("supabase token missing email claim: %w", domain.ErrUnauthenticated)
	}

	if !readEmailVerified(claims) {
		return "", "", fmt.Errorf("supabase token sub %s: %w", sub, domain.ErrSupabaseEmailNotVerified)
	}

	return sub, email, nil
}

// readEmailVerified checks email_verified at top level and in user_metadata (SDKs differ).
func readEmailVerified(claims jwt.MapClaims) bool {
	if v, ok := claims["email_verified"].(bool); ok && v {
		return true
	}
	meta, ok := claims["user_metadata"].(map[string]any)
	if !ok {
		return false
	}
	v, ok := meta["email_verified"].(bool)
	return ok && v
}

// Close shuts down the background JWKS refresh goroutine.
func (v *SupabaseVerifier) Close() {
	v.cancel()
}
