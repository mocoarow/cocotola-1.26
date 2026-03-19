package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_NewTokenHash_shouldReturnHash_whenValidHex64(t *testing.T) {
	t.Parallel()

	// given
	validHex := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	// when
	hash, err := domain.NewTokenHash(validHex)

	// then
	require.NoError(t, err)
	assert.Equal(t, domain.TokenHash(validHex), hash)
}

func Test_NewTokenHash_shouldReturnError_whenLengthIsNot64(t *testing.T) {
	t.Parallel()

	// given
	shortHex := "abcdef"

	// when
	_, err := domain.NewTokenHash(shortHex)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "64 characters")
}

func Test_NewTokenHash_shouldReturnError_whenNotValidHex(t *testing.T) {
	t.Parallel()

	// given
	invalidHex := "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"

	// when
	_, err := domain.NewTokenHash(invalidHex)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "valid hex")
}

func Test_NewLoginID_shouldReturnLoginID_whenNotEmpty(t *testing.T) {
	t.Parallel()

	// given
	id := "user@example.com"

	// when
	loginID, err := domain.NewLoginID(id)

	// then
	require.NoError(t, err)
	assert.Equal(t, domain.LoginID(id), loginID)
}

func Test_NewLoginID_shouldReturnError_whenEmpty(t *testing.T) {
	t.Parallel()

	// given / when
	_, err := domain.NewLoginID("")

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func Test_TokenHash_String_shouldReturnStringRepresentation(t *testing.T) {
	t.Parallel()

	// given
	hex := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	hash := domain.TokenHash(hex)

	// when
	result := hash.String()

	// then
	assert.Equal(t, hex, result)
}

func Test_LoginID_String_shouldReturnStringRepresentation(t *testing.T) {
	t.Parallel()

	// given
	id := domain.LoginID("user@example.com")

	// when
	result := id.String()

	// then
	assert.Equal(t, "user@example.com", result)
}
