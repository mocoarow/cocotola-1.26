package gateway

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// SupabaseVerifier verifies Supabase JWT tokens using the shared JWT secret.
type SupabaseVerifier struct {
	secret []byte
}

// NewSupabaseVerifier returns a new SupabaseVerifier.
func NewSupabaseVerifier(jwtSecret string) *SupabaseVerifier {
	return &SupabaseVerifier{secret: []byte(jwtSecret)}
}

// Verify parses and validates a Supabase JWT, returning the user's sub (UUID) and email.
func (v *SupabaseVerifier) Verify(_ context.Context, tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})
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
