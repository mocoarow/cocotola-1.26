package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
)

type stubUserSaver struct {
	err error
}

func (s *stubUserSaver) Save(_ context.Context, _ *user.AppUser) error {
	return s.err
}

func Test_Provision_shouldReturnAppUser_whenSaverSucceeds(t *testing.T) {
	t.Parallel()

	// given
	saver := &stubUserSaver{err: nil}

	// when
	u, err := user.Provision(context.Background(), saver, fixtureOrgID, "user@example.com", "$2a$10$hashedpassword", true)

	// then
	require.NoError(t, err)
	assert.False(t, u.ID().IsZero())
	assert.Equal(t, fixtureOrgID, u.OrganizationID())
	assert.Equal(t, domain.LoginID("user@example.com"), u.LoginID())
	assert.Equal(t, "$2a$10$hashedpassword", u.HashedPassword())
	assert.True(t, u.Enabled())
}

func Test_Provision_shouldReturnError_whenSaverFails(t *testing.T) {
	t.Parallel()

	// given
	saverErr := errors.New("db error")
	saver := &stubUserSaver{err: saverErr}

	// when
	_, err := user.Provision(context.Background(), saver, fixtureOrgID, "user@example.com", "$2a$10$hashedpassword", true)

	// then
	require.ErrorIs(t, err, saverErr)
}

func Test_Provision_shouldReturnError_whenLoginIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	saver := &stubUserSaver{}

	// when
	_, err := user.Provision(context.Background(), saver, fixtureOrgID, "", "$2a$10$hashedpassword", true)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
