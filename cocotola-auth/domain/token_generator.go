package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const opaqueTokenBytes = 32

// GenerateOpaqueToken generates a cryptographically random opaque token and its SHA256 hash.
// Returns (raw token, token hash, error).
func GenerateOpaqueToken() (string, TokenHash, error) {
	b := make([]byte, opaqueTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate random bytes: %w", err)
	}
	raw := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	hash := HashToken(raw)
	return raw, hash, nil
}

// HashToken returns the SHA256 hex digest of the given raw token string.
func HashToken(raw string) TokenHash {
	h := sha256.Sum256([]byte(raw))
	return TokenHash(hex.EncodeToString(h[:]))
}
