package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
)

var (
	fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")
	fixtureOrgID     = domain.MustParseOrganizationID("00000000-0000-7000-8000-000000000010")
)

func validAppUserArgs() (domain.AppUserID, domain.OrganizationID, domain.LoginID, string, bool) {
	return fixtureAppUserID, fixtureOrgID, "user@example.com", "$2a$10$hashedpassword", true
}

func Test_NewAppUser_shouldReturnAppUser_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	u, err := user.NewAppUser(id, orgID, loginID, hashedPw, enabled)

	// then
	require.NoError(t, err)
	assert.True(t, id.Equal(u.ID()))
	assert.True(t, orgID.Equal(u.OrganizationID()))
	assert.Equal(t, loginID, u.LoginID())
	assert.Equal(t, hashedPw, u.HashedPassword())
	assert.True(t, u.Enabled())
}

func Test_NewAppUser_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(domain.AppUserID{}, orgID, loginID, hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(id, domain.OrganizationID{}, loginID, hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenLoginIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, hashedPw, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(id, orgID, "", hashedPw, enabled)

	// then
	require.Error(t, err)
}

func Test_AppUser_Enable_shouldSetEnabledTrue_whenDisabled(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(fixtureAppUserID, fixtureOrgID, "user@example.com", "$2a$10$hash", false)

	// when
	u.Enable()

	// then
	assert.True(t, u.Enabled())
}

func Test_AppUser_Disable_shouldSetEnabledFalse_whenEnabled(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(fixtureAppUserID, fixtureOrgID, "user@example.com", "$2a$10$hash", true)

	// when
	u.Disable()

	// then
	assert.False(t, u.Enabled())
}
