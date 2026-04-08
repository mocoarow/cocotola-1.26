package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
)

func validAppUserArgs() (int, int, domain.LoginID, string, string, string, bool) {
	return 1, 1, "user@example.com", "$2a$10$hashedpassword", "", "", true
}

func Test_NewAppUser_shouldReturnAppUser_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, loginID, hashedPw, provider, providerID, enabled := validAppUserArgs()

	// when
	u, err := user.NewAppUser(id, orgID, loginID, hashedPw, provider, providerID, enabled)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, u.ID())
	assert.Equal(t, orgID, u.OrganizationID())
	assert.Equal(t, loginID, u.LoginID())
	assert.Equal(t, hashedPw, u.HashedPassword())
	assert.False(t, u.IsLinkedToProvider())
	assert.True(t, u.Enabled())
}

func Test_NewAppUser_shouldReturnError_whenIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, loginID, hashedPw, provider, providerID, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(0, orgID, loginID, hashedPw, provider, providerID, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenIDIsNegative(t *testing.T) {
	t.Parallel()

	// given
	_, orgID, loginID, hashedPw, provider, providerID, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(-1, orgID, loginID, hashedPw, provider, providerID, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenOrganizationIDIsZero(t *testing.T) {
	t.Parallel()

	// given
	id, _, loginID, hashedPw, provider, providerID, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(id, 0, loginID, hashedPw, provider, providerID, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenLoginIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, orgID, _, hashedPw, provider, providerID, enabled := validAppUserArgs()

	// when
	_, err := user.NewAppUser(id, orgID, "", hashedPw, provider, providerID, enabled)

	// then
	require.Error(t, err)
}

func Test_NewAppUser_shouldReturnError_whenProviderSetButProviderIDEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "supabase", "", true)

	// then
	require.Error(t, err)
}

func Test_AppUser_Enable_shouldSetEnabledTrue_whenDisabled(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "", "", false)

	// when
	u.Enable()

	// then
	assert.True(t, u.Enabled())
}

func Test_AppUser_Disable_shouldSetEnabledFalse_whenEnabled(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "", "", true)

	// when
	u.Disable()

	// then
	assert.False(t, u.Enabled())
}

func Test_AppUser_LinkProvider_shouldSetProvider_whenUserHasNoProvider(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "", "", true)

	// when
	err := u.LinkProvider("supabase", "sub-123")

	// then
	require.NoError(t, err)
	assert.True(t, u.IsLinkedToProvider())
	assert.Equal(t, "supabase", u.Provider())
	assert.Equal(t, "sub-123", u.ProviderID())
}

func Test_AppUser_LinkProvider_shouldReturnError_whenUserAlreadyLinked(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "supabase", "sub-existing", true)

	// when
	err := u.LinkProvider("supabase", "sub-other")

	// then
	require.ErrorIs(t, err, domain.ErrAppUserAlreadyLinked)
	assert.Equal(t, "sub-existing", u.ProviderID())
}

func Test_AppUser_LinkProvider_shouldReturnError_whenProviderEmpty(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "", "", true)

	// when
	err := u.LinkProvider("", "sub-123")

	// then
	require.Error(t, err)
}

func Test_AppUser_LinkProvider_shouldReturnError_whenProviderIDEmpty(t *testing.T) {
	t.Parallel()

	// given
	u, _ := user.NewAppUser(1, 1, "user@example.com", "$2a$10$hash", "", "", true)

	// when
	err := u.LinkProvider("supabase", "")

	// then
	require.Error(t, err)
}
