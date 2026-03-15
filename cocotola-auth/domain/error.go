package domain

import "errors"

// ErrUnauthenticated is returned when authentication credentials are invalid or missing.
var ErrUnauthenticated = errors.New("unauthenticated")

// ErrTokenNotFound is returned when a token cannot be found in the store.
var ErrTokenNotFound = errors.New("token not found")

// ErrTokenRevoked is returned when a revoked token is used.
var ErrTokenRevoked = errors.New("token revoked")

// ErrSessionExpired is returned when a session token has exceeded its absolute timeout.
var ErrSessionExpired = errors.New("session expired")
