package domain

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// MinPasswordLength is the minimum number of characters (runes) for a password.
const MinPasswordLength = 8

// ErrPasswordTooShort is returned when a password is shorter than MinPasswordLength.
var ErrPasswordTooShort = errors.New("password too short")

// PasswordHasher hashes and compares passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword string, password string) error
}

// HashPassword validates the raw password against the policy and returns the hashed result.
func HashPassword(rawPassword string, hasher PasswordHasher) (string, error) {
	if utf8.RuneCountInString(rawPassword) < MinPasswordLength {
		return "", ErrPasswordTooShort
	}
	hashed, err := hasher.Hash(rawPassword)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return hashed, nil
}
