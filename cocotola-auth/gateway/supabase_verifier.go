package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

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

// Verify parses and validates a Supabase JWT, returning the user's sub (UUID) and email.
func (v *SupabaseVerifier) Verify(ctx context.Context, tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, v.jwks.KeyfuncCtx(ctx))
	if err != nil {
		return "", "", fmt.Errorf("parse supabase token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("invalid supabase token claims")
	}

	sub, _ := claims.GetSubject()
	if sub == "" {
		return "", "", errors.New("supabase token missing sub claim")
	}

	email, _ := claims["email"].(string)
	if email == "" {
		return "", "", errors.New("supabase token missing email claim")
	}

	return sub, email, nil
}

// Close shuts down the background JWKS refresh goroutine.
func (v *SupabaseVerifier) Close() {
	v.cancel()
}
