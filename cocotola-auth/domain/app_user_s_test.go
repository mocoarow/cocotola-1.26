package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func validAppUserArgs() (int, int, domain.LoginID, string, bool) {
	return 1, 1, "user@example.com", "$2a$10$hashedpassword", true
}

func Test_NewAppUser_shouldReturnAppUser_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	user, err := domain.NewAppUser(id, orgID, loginID, hashedPw, enabled)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, user.ID())
	assert.Equal(t, orgID, user.OrganizationID())
	assert.Equal(t, loginID, user.LoginID())
	assert.Equal(t, hashedPw, user.HashedPassword())
	assert.True(t, user.Enabled())
}

func Test_NewAppUser_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := domain.NewAppUser(0, orgID, loginID, hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenIDIsNegative(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := domain.NewAppUser(-1, orgID, loginID, hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := domain.NewAppUser(id, 0, loginID, hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenLoginIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := domain.NewAppUser(id, orgID, "", hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_AppUser_Enable_shouldSetEnabledTrue_whenDisabled(t *testing.T) {
	t.Parallel()

	// given
	user, _ := domain.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", false)

	// when
	user.Enable()

	// then
	assert.True(t, user.Enabled())
}

func Test_AppUser_Disable_shouldSetEnabledFalse_whenEnabled(t *testing.T) {
	t.Parallel()

	// given
	user, _ := domain.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", true)

	// when
	user.Disable()

	// then
	assert.False(t, user.Enabled())
}
