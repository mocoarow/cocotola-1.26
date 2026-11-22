package gateway_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func Test_BcryptHasher_Compare_shouldReturnNil_whenPasswordMatchesHash(t *testing.T) {
	t.Parallel()

	// given
	hasher := gateway.NewBcryptHasher()
	hashed, err := hasher.Hash("correct-password")
	require.NoError(t, err)

	// when
	err = hasher.Compare(hashed, "correct-password")

	// then
	require.NoError(t, err)
}

func Test_BcryptHasher_Compare_shouldReturnErrMismatchedHashAndPassword_whenPasswordDiffers(t *testing.T) {
	t.Parallel()

	// given
	hasher := gateway.NewBcryptHasher()
	hashed, err := hasher.Hash("correct-password")
	require.NoError(t, err)

	// when
	err = hasher.Compare(hashed, "wrong-password")

	// then
	require.ErrorIs(t, err, bcrypt.ErrMismatchedHashAndPassword)
}
