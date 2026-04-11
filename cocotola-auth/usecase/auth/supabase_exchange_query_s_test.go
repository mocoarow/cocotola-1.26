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

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnExistingUser_whenProviderLinkExists(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-123", "user@example.com", nil)

	providerLink := domainuser.ReconstructAppUserProvider(
		domain.MustParseAppUserProviderID("00000000-0000-7000-8000-000000000060"),
		fixtureSupaUserID1, fixtureOrgID, "supabase", "sub-123",
	)
	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-123").Return(providerLink, nil)

	providerSaverMock := NewMockAppUserProviderSaver(t)

	appUser := domainuser.ReconstructAppUser(fixtureSupaUserID1, fixtureOrgID, "user@example.com", "", true)
	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)
	appUserByIDFinderMock.On("FindByID", mock.Anything, fixtureSupaUserID1).Return(appUser, nil)

	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	saverMock := NewMockAppUserSaver(t)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
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

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldCreateUserAndLink_whenUserDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-456", "new@example.com", nil)

	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-456").
		Return(nil, domain.ErrAppUserProviderNotFound)

	providerSaverMock := NewMockAppUserProviderSaver(t)
	providerSaverMock.On("Save", mock.Anything, mock.MatchedBy(func(p *domainuser.AppUserProvider) bool {
		return p.Provider() == "supabase" && p.ProviderID() == "sub-456"
	})).Return(nil)

	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.MatchedBy(func(u *domainuser.AppUser) bool {
		return !u.ID().IsZero() && string(u.LoginID()) == "new@example.com"
	})).Return(nil)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
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

	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerSaverMock := NewMockAppUserProviderSaver(t)
	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	saverMock := NewMockAppUserSaver(t)
	orgFinderMock := NewMockOrganizationFinder(t)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
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

	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerSaverMock := NewMockAppUserProviderSaver(t)
	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	saverMock := NewMockAppUserSaver(t)

	orgFinderMock := NewMockOrganizationFinder(t)
	orgFinderMock.On("FindByName", mock.Anything, "unknown-org").Return(nil, domain.ErrOrganizationNotFound)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
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

	providerLink := domainuser.ReconstructAppUserProvider(
		domain.MustParseAppUserProviderID("00000000-0000-7000-8000-000000000061"),
		fixtureSupaUserID2, fixtureOrgID, "supabase", "sub-789",
	)
	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-789").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-789").
		Return(providerLink, nil).Once()

	providerSaverMock := NewMockAppUserProviderSaver(t)

	appUser := domainuser.ReconstructAppUser(fixtureSupaUserID2, fixtureOrgID, "race@example.com", "", true)
	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)
	appUserByIDFinderMock.On("FindByID", mock.Anything, fixtureSupaUserID2).Return(appUser, nil)

	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.True(t, fixtureSupaUserID2.Equal(output.UserID))
	assert.Equal(t, "race@example.com", output.LoginID)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldLinkProvider_whenUserExistsByLoginIDWithoutPassword(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-999", "existing@example.com", nil)

	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-999").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-999").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()

	providerSaverMock := NewMockAppUserProviderSaver(t)
	providerSaverMock.On("Save", mock.Anything, mock.MatchedBy(func(p *domainuser.AppUserProvider) bool {
		return p.AppUserID().Equal(fixtureSupaUserID3) && p.Provider() == "supabase" && p.ProviderID() == "sub-999"
	})).Return(nil)

	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)

	existingUser := domainuser.ReconstructAppUser(fixtureSupaUserID3, fixtureOrgID, "existing@example.com", "", true)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	loginIDFinderMock.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID("existing@example.com")).Return(existingUser, nil)

	saverMock := NewMockAppUserSaver(t)
	// First Save attempt for the new aggregate fails (duplicate login_id).
	saverMock.On("Save", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
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

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-new", "human@example.com", nil)

	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-new").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-new").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()

	providerSaverMock := NewMockAppUserProviderSaver(t)
	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)

	passwordAccount := domainuser.ReconstructAppUser(fixtureSupaUserID1, fixtureOrgID, "human@example.com", "$2a$10$hash", true)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)
	loginIDFinderMock.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID("human@example.com")).Return(passwordAccount, nil)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrAppUserAutoLinkRejected)
	require.Nil(t, output)
	saverMock.AssertNumberOfCalls(t, "Save", 1)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldPropagateSaveError_whenCreateFailsWithNonDuplicateError(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-oops", "oops@example.com", nil)

	providerFinderMock := NewMockAppUserProviderFinder(t)
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-oops").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()
	providerFinderMock.On("FindByProviderID", mock.Anything, fixtureOrgID, "supabase", "sub-oops").
		Return(nil, domain.ErrAppUserProviderNotFound).Once()

	providerSaverMock := NewMockAppUserProviderSaver(t)
	appUserByIDFinderMock := NewMockAppUserByIDFinder(t)
	loginIDFinderMock := NewMockAppUserByLoginIDFinder(t)

	saverMock := NewMockAppUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(errors.New("db unavailable")).Once()

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, providerFinderMock, providerSaverMock, appUserByIDFinderMock, loginIDFinderMock, saverMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.Error(t, err)
	require.Nil(t, output)
	assert.Contains(t, err.Error(), "db unavailable")
	loginIDFinderMock.AssertNotCalled(t, "FindByLoginID")
}
