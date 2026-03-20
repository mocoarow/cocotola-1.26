package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type stubHasher struct {
	hashResult string
	hashErr    error
	compareErr error
}

func (s *stubHasher) Hash(_ string) (string, error) {
	return s.hashResult, s.hashErr
}

func (s *stubHasher) Compare(_, _ string) error {
	return s.compareErr
}

func Test_AppUser_ChangePassword_shouldUpdateHash_whenPasswordIsValid(t *testing.T) {
	t.Parallel()

	// given
	user := domain.ReconstructAppUser(1, 1, "user@example.com", "", true)
	hasher := &stubHasher{hashResult: "$2a$10$newhash"}

	// when
	err := user.ChangePassword("validpass", hasher)

	// then
	require.NoError(t, err)
	assert.Equal(t, "$2a$10$newhash", user.HashedPassword())
}

func Test_AppUser_ChangePassword_shouldReturnError_whenPasswordTooShort(t *testing.T) {
	t.Parallel()

	// given
	user := domain.ReconstructAppUser(1, 1, "user@example.com", "", true)
	hasher := &stubHasher{}

	// when
	err := user.ChangePassword("short", hasher)

	// then
	require.ErrorIs(t, err, domain.ErrPasswordTooShort)
}

func Test_AppUser_ChangePassword_shouldReturnError_whenHasherFails(t *testing.T) {
	t.Parallel()

	// given
	user := domain.ReconstructAppUser(1, 1, "user@example.com", "", true)
	hashErr := errors.New("hash failure")
	hasher := &stubHasher{hashErr: hashErr}

	// when
	err := user.ChangePassword("validpass", hasher)

	// then
	require.ErrorIs(t, err, hashErr)
}

func Test_AppUser_VerifyPassword_shouldReturnNil_whenPasswordMatches(t *testing.T) {
	t.Parallel()

	// given
	user := domain.ReconstructAppUser(1, 1, "user@example.com", "$2a$10$hash", true)
	hasher := &stubHasher{compareErr: nil}

	// when
	err := user.VerifyPassword("correct", hasher)

	// then
	require.NoError(t, err)
}

func Test_AppUser_VerifyPassword_shouldReturnError_whenPasswordDoesNotMatch(t *testing.T) {
	t.Parallel()

	// given
	user := domain.ReconstructAppUser(1, 1, "user@example.com", "$2a$10$hash", true)
	compareErr := errors.New("mismatch")
	hasher := &stubHasher{compareErr: compareErr}

	// when
	err := user.VerifyPassword("wrong", hasher)

	// then
	require.ErrorIs(t, err, compareErr)
}
