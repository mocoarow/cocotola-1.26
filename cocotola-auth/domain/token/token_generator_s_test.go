package token_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

func Test_GenerateOpaqueToken_shouldReturnNonEmptyTokenAndHash(t *testing.T) {
	t.Parallel()

	// given / when
	raw, hash, err := token.GenerateOpaqueToken()

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.Len(t, string(hash), 64)
}

func Test_GenerateOpaqueToken_shouldReturnValidHexHash(t *testing.T) {
	t.Parallel()

	// given / when
	_, hash, err := token.GenerateOpaqueToken()

	// then
	require.NoError(t, err)
	_, decodeErr := hex.DecodeString(string(hash))
	require.NoError(t, decodeErr)
}

func Test_GenerateOpaqueToken_shouldReturnHashMatchingRawToken(t *testing.T) {
	t.Parallel()

	// given / when
	raw, hash, err := token.GenerateOpaqueToken()

	// then
	require.NoError(t, err)
	expected := token.HashToken(raw)
	assert.Equal(t, expected, hash)
}

func Test_GenerateOpaqueToken_shouldReturnDifferentTokensEachCall(t *testing.T) {
	t.Parallel()

	// given / when
	raw1, _, _ := token.GenerateOpaqueToken()
	raw2, _, _ := token.GenerateOpaqueToken()

	// then
	assert.NotEqual(t, raw1, raw2)
}

func Test_HashToken_shouldReturn64CharHexDigest(t *testing.T) {
	t.Parallel()

	// given
	raw := "test-token-string"

	// when
	hash := token.HashToken(raw)

	// then
	assert.Len(t, string(hash), 64)
	_, err := hex.DecodeString(string(hash))
	require.NoError(t, err)
}

func Test_HashToken_shouldReturnSameHashForSameInput(t *testing.T) {
	t.Parallel()

	// given
	raw := "deterministic-input"

	// when
	hash1 := token.HashToken(raw)
	hash2 := token.HashToken(raw)

	// then
	assert.Equal(t, hash1, hash2)
}

func Test_HashToken_shouldReturnDifferentHashForDifferentInput(t *testing.T) {
	t.Parallel()

	// given / when
	hash1 := token.HashToken("input-a")
	hash2 := token.HashToken("input-b")

	// then
	assert.NotEqual(t, hash1, hash2)
}
