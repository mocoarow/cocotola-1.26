package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

var (
	fixtureSupaUserID1 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000051")
	fixtureSupaUserID2 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000052")
	fixtureSupaUserID3 = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000053")
)

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnExistingUser_whenUserExists(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-123", "user@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	appUser := domainuser.ReconstructAppUser(fixtureSupaUserID1, fixtureOrgID, "user@example.com", "", "supabase", "sub-123", true)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-123").Return(appUser, nil)

	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	saverMock := NewMockAppUserSaver(t)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.True(t, fixtureSupaUserID1.Equal(output.UserID))
	assert.Equal(t, "user@example.com", output.LoginID)
	assert.Equal(t, "test-org", output.OrganizationName)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldCreateUser_whenUserDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-456", "new@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-456").
		Return(nil, domain.ErrAppUserNotFound)

	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.MatchedBy(func(u *domainuser.AppUser) bool {
		return !u.ID().IsZero() &&
			string(u.LoginID()) == "new@example.com" &&
			u.Provider() == "supabase" &&
			u.ProviderID() == "sub-456"
	})).Return(nil)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.False(t, output.UserID.IsZero())
	assert.Equal(t, "new@example.com", output.LoginID)
	assert.Equal(t, "test-org", output.OrganizationName)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnError_whenTokenIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "bad-jwt").Return("", "", errors.New("invalid token"))

	finderMock := NewMockAppUserProviderFinder(t)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	saverMock := NewMockAppUserSaver(t)
	orgFinderMock := NewMockOrganizationFinder(t)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("bad-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.Error(t, err)
	require.Nil(t, output)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-123", "user@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	saverMock := NewMockAppUserSaver(t)

	orgFinderMock := NewMockOrganizationFinder(t)
	orgFinderMock.On("FindByName", mock.Anything, "unknown-org").Return(nil, domain.ErrOrganizationNotFound)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "unknown-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	require.Nil(t, output)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldRetryFind_whenCreateRaceCondition(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-789", "race@example.com", nil)

	appUser := domainuser.ReconstructAppUser(fixtureSupaUserID2, fixtureOrgID, "race@example.com", "", "supabase", "sub-789", true)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-789").
		Return(nil, domain.ErrAppUserNotFound).Once()
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-789").
		Return(appUser, nil).Once()

	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.True(t, fixtureSupaUserID2.Equal(output.UserID))
	assert.Equal(t, "race@example.com", output.LoginID)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldLinkProvider_whenUserExistsByLoginIDWithoutProvider(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-999", "existing@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-999").
		Return(nil, domain.ErrAppUserNotFound).Once()
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-999").
		Return(nil, domain.ErrAppUserNotFound).Once()

	// Existing passwordless user with no provider yet — only these are eligible
	// for auto-linking; password-holding accounts MUST be rejected (see C1).
	existingUser := domainuser.ReconstructAppUser(fixtureSupaUserID3, fixtureOrgID, "existing@example.com", "", "", "", true)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	loginIDFinderMock.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID("existing@example.com")).Return(existingUser, nil)

	saverMock := NewMockAppUserSaver(t)
	// First Save attempt for the new aggregate fails (duplicate login_id).
	saverMock.On("Save", mock.Anything, mock.MatchedBy(func(u *domainuser.AppUser) bool {
		return !u.ID().Equal(fixtureSupaUserID3)
	})).Return(gorm.ErrDuplicatedKey).Once()
	// Second Save persists the linked existing aggregate.
	saverMock.On("Save", mock.Anything, mock.MatchedBy(func(u *domainuser.AppUser) bool {
		return u.ID().Equal(fixtureSupaUserID3) && u.Provider() == "supabase" && u.ProviderID() == "sub-999"
	})).Return(nil).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.True(t, fixtureSupaUserID3.Equal(output.UserID))
	assert.Equal(t, "existing@example.com", output.LoginID)
	assert.Equal(t, "test-org", output.OrganizationName)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldRejectLink_whenExistingAccountHasPassword(t *testing.T) {
	t.Parallel()

	// given: an existing local password account exists for the same email. SupabaseVerifier
	// already confirmed email_verified=true, but auto-linking a password-holding account
	// would enable account takeover, so the exchange MUST refuse.
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-new", "human@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-new").
		Return(nil, domain.ErrAppUserNotFound).Once()
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-new").
		Return(nil, domain.ErrAppUserNotFound).Once()

	passwordAccount := domainuser.ReconstructAppUser(fixtureSupaUserID1, fixtureOrgID, "human@example.com", "$2a$10$hash", "", "", true)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	loginIDFinderMock.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID("human@example.com")).Return(passwordAccount, nil)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrAppUserAutoLinkRejected)
	require.Nil(t, output)
	// The password account must NEVER be re-saved.
	saverMock.AssertNumberOfCalls(t, "Save", 1)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldPropagateSaveError_whenCreateFailsWithNonDuplicateError(t *testing.T) {
	t.Parallel()

	// given: a non-duplicate error (e.g. network/DB fault) from Save must propagate
	// — we must NOT enter the linking branch and silently paper over persistence bugs.
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-oops", "oops@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-oops").
		Return(nil, domain.ErrAppUserNotFound).Once()
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-oops").
		Return(nil, domain.ErrAppUserNotFound).Once()

	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(errors.New("db unavailable")).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.Error(t, err)
	require.Nil(t, output)
	assert.Contains(t, err.Error(), "db unavailable")
	// FindByLoginID must NOT be called because we did not detect a duplicate key.
	loginIDFinderMock.AssertNotCalled(t, "FindByLoginID")
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnError_whenExistingUserAlreadyLinkedToAnotherProvider(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-attacker", "victim@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-attacker").
		Return(nil, domain.ErrAppUserNotFound).Once()
	finderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-attacker").
		Return(nil, domain.ErrAppUserNotFound).Once()

	// Victim is already linked to a different provider id; aggregate must reject relinking.
	victim := domainuser.ReconstructAppUser(fixtureSupaUserID2, fixtureOrgID, "victim@example.com", "", "supabase", "sub-victim", true)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	loginIDFinderMock.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID("victim@example.com")).Return(victim, nil)

	saverMock := NewMockAppUserSaver(t)
	// First Save (new aggregate) fails because login_id collides.
	saverMock.On("Save", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrAppUserAlreadyLinked)
	require.Nil(t, output)
	// The victim aggregate must never be re-saved.
	saverMock.AssertNumberOfCalls(t, "Save", 1)
}
