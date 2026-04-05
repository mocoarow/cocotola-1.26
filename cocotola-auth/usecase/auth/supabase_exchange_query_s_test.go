package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnExistingUser_whenUserExists(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-123", "user@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	appUser := domainuser.ReconstructAppUser(10, 1, "user@example.com", "", true)
	finderMock.On("FindByProviderID", mock.Anything, 1, "supabase", "sub-123").Return(appUser, nil)

	creatorMock := NewMockAppUserProviderCreator(t)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(1, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, creatorMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 10, output.UserID)
	assert.Equal(t, "user@example.com", output.LoginID)
	assert.Equal(t, "test-org", output.OrganizationName)
	creatorMock.AssertNotCalled(t, "CreateWithProvider")
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldCreateUser_whenUserDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-456", "new@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, 1, "supabase", "sub-456").Return(nil, domain.ErrAppUserNotFound)

	creatorMock := NewMockAppUserProviderCreator(t)
	creatorMock.On("CreateWithProvider", mock.Anything, 1, "new@example.com", "supabase", "sub-456").Return(20, nil)

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(1, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, creatorMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 20, output.UserID)
	assert.Equal(t, "new@example.com", output.LoginID)
	assert.Equal(t, "test-org", output.OrganizationName)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnError_whenTokenIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "bad-jwt").Return("", "", errors.New("invalid token"))

	finderMock := NewMockAppUserProviderFinder(t)
	creatorMock := NewMockAppUserProviderCreator(t)
	orgFinderMock := NewMockOrganizationFinder(t)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, creatorMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("bad-jwt", "test-org")
	require.NoError(t, err)

	// when
	_, err = query.SupabaseExchange(context.Background(), input)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "verify supabase token")
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-123", "user@example.com", nil)

	finderMock := NewMockAppUserProviderFinder(t)
	creatorMock := NewMockAppUserProviderCreator(t)

	orgFinderMock := NewMockOrganizationFinder(t)
	orgFinderMock.On("FindByName", mock.Anything, "unknown-org").Return(nil, domain.ErrOrganizationNotFound)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, creatorMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "unknown-org")
	require.NoError(t, err)

	// when
	_, err = query.SupabaseExchange(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
}

func Test_SupabaseExchangeQuery_SupabaseExchange_shouldRetryFind_whenCreateRaceCondition(t *testing.T) {
	t.Parallel()

	// given
	verifierMock := NewMockSupabaseVerifier(t)
	verifierMock.On("Verify", mock.Anything, "supabase-jwt").Return("sub-789", "race@example.com", nil)

	appUser := domainuser.ReconstructAppUser(30, 1, "race@example.com", "", true)

	finderMock := NewMockAppUserProviderFinder(t)
	finderMock.On("FindByProviderID", mock.Anything, 1, "supabase", "sub-789").
		Return(nil, domain.ErrAppUserNotFound).Once()
	finderMock.On("FindByProviderID", mock.Anything, 1, "supabase", "sub-789").
		Return(appUser, nil).Once()

	creatorMock := NewMockAppUserProviderCreator(t)
	creatorMock.On("CreateWithProvider", mock.Anything, 1, "race@example.com", "supabase", "sub-789").
		Return(0, errors.New("duplicate key"))

	orgFinderMock := NewMockOrganizationFinder(t)
	org := domain.ReconstructOrganization(1, "test-org", 100, 50)
	orgFinderMock.On("FindByName", mock.Anything, "test-org").Return(org, nil)

	query := authusecase.NewSupabaseExchangeQuery(verifierMock, finderMock, creatorMock, orgFinderMock)
	input, err := authservice.NewSupabaseExchangeInput("supabase-jwt", "test-org")
	require.NoError(t, err)

	// when
	output, err := query.SupabaseExchange(context.Background(), input)

	// then
	require.NoError(t, err)
	assert.Equal(t, 30, output.UserID)
	assert.Equal(t, "race@example.com", output.LoginID)
}
