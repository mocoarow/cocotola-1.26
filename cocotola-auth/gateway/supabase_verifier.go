package gateway

import (
	"context"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// allowedSupabaseAlgs returns the set of JWT signing algorithms accepted from
// Supabase. Asymmetric algorithms only — HS* family is explicitly rejected so a
// compromised HS key cannot be used to forge tokens, and "none" is rejected so
// unsigned tokens are never accepted.
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

// Verify parses and validates a Supabase JWT against the pinned allow-list of
// signing algorithms, and returns the verified sub (UUID) and email.
//
// Verify returns domain.ErrSupabaseEmailNotVerified when the token is
// cryptographically valid but email_verified is missing or false. Callers MUST
// NOT treat such a token as trustworthy: an unverified email can be spoofed
// and would enable account takeover via the auto-linking path.
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
		return "", "", fmt.Errorf("supabase token email=%s: %w", email, domain.ErrSupabaseEmailNotVerified)
	}

	return sub, email, nil
}

// readEmailVerified extracts the email_verified claim, checking both the
// top-level claim (Supabase projects surface it at top-level when the user is
// confirmed) and user_metadata.email_verified (some SDKs place it there).
// Any value that is not exactly the boolean true results in false.
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
