package domain

import (
	"encoding/hex"
	"fmt"
)

// TokenHashLength is the expected length of a SHA256 hex digest.
const TokenHashLength = 64

// TokenHash represents a SHA256 hex digest of a raw token.
type TokenHash string

// NewTokenHash creates a validated TokenHash.
func NewTokenHash(hash string) (TokenHash, error) {
	if len(hash) != TokenHashLength {
		return "", fmt.Errorf("token hash must be %d characters, got %d: %w", TokenHashLength, len(hash), ErrInvalidArgument)
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return "", fmt.Errorf("token hash must be valid hex: %w", ErrInvalidArgument)
	}
	return TokenHash(hash), nil
}

// String returns the string representation.
func (h TokenHash) String() string {
	return string(h)
}

// LoginID represents a user's login identifier.
type LoginID string

// NewLoginID creates a validated LoginID.
func NewLoginID(id string) (LoginID, error) {
	if id == "" {
		return "", fmt.Errorf("login id must not be empty: %w", ErrInvalidArgument)
	}
	return LoginID(id), nil
}

// String returns the string representation.
func (id LoginID) String() string {
	return string(id)
}
