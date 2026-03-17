package domain_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_GenerateOpaqueToken_shouldReturnNonEmptyTokenAndHash(t *testing.T) {
	t.Parallel()

	// given / when
	raw, hash, err := domain.GenerateOpaqueToken()

	// then
	assert.NoError(t, err)
	assert.NotEmpty(t, raw)
	assert.Len(t, string(hash), 64)
}

func Test_GenerateOpaqueToken_shouldReturnValidHexHash(t *testing.T) {
	t.Parallel()

	// given / when
	_, hash, err := domain.GenerateOpaqueToken()

	// then
	assert.NoError(t, err)
	_, decodeErr := hex.DecodeString(string(hash))
	assert.NoError(t, decodeErr)
}

func Test_GenerateOpaqueToken_shouldReturnHashMatchingRawToken(t *testing.T) {
	t.Parallel()

	// given / when
	raw, hash, err := domain.GenerateOpaqueToken()

	// then
	assert.NoError(t, err)
	expected := domain.HashToken(raw)
	assert.Equal(t, expected, hash)
}

func Test_GenerateOpaqueToken_shouldReturnDifferentTokensEachCall(t *testing.T) {
	t.Parallel()

	// given / when
	raw1, _, _ := domain.GenerateOpaqueToken()
	raw2, _, _ := domain.GenerateOpaqueToken()

	// then
	assert.NotEqual(t, raw1, raw2)
}

func Test_HashToken_shouldReturn64CharHexDigest(t *testing.T) {
	t.Parallel()

	// given
	raw := "test-token-string"

	// when
	hash := domain.HashToken(raw)

	// then
	assert.Len(t, string(hash), 64)
	_, err := hex.DecodeString(string(hash))
	assert.NoError(t, err)
}

func Test_HashToken_shouldReturnSameHashForSameInput(t *testing.T) {
	t.Parallel()

	// given
	raw := "deterministic-input"

	// when
	hash1 := domain.HashToken(raw)
	hash2 := domain.HashToken(raw)

	// then
	assert.Equal(t, hash1, hash2)
}

func Test_HashToken_shouldReturnDifferentHashForDifferentInput(t *testing.T) {
	t.Parallel()

	// given / when
	hash1 := domain.HashToken("input-a")
	hash2 := domain.HashToken("input-b")

	// then
	assert.NotEqual(t, hash1, hash2)
}
