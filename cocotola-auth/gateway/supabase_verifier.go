package gateway

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// allowedSupabaseAlgs is asymmetric-only: HS* and "none" must never be accepted.
func allowedSupabaseAlgs() []string {
	return []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}
}

// SupabaseVerifier verifies Supabase JWT tokens using JWKS.
type SupabaseVerifier struct {
	jwks   keyfunc.Keyfunc
	cancel context.CancelFunc
}

// NewSupabaseVerifier fetches the JWKS from the given URL and returns a new SupabaseVerifier.
func NewSupabaseVerifier(ctx context.Context, jwksURL string) (*SupabaseVerifier, error) {
	ctx, cancel := context.WithCancel(ctx)

	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("fetch JWKS from %s: %w", jwksURL, err)
	}

	return &SupabaseVerifier{jwks: jwks, cancel: cancel}, nil
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
