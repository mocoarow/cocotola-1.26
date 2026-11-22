package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

const fixtureGuestOrgName = "guest-auth-org"

var fixtureGuestLoginID = domain.LoginID(domainuser.NewGuestLoginID(fixtureGuestOrgName))

func newGuestAuthOrgFinder(t *testing.T) *MockOrganizationFinder {
	t.Helper()
	org := domain.ReconstructOrganization(fixtureOrgID, fixtureGuestOrgName, 100, 50)
	orgFinder := NewMockOrganizationFinder(t)
	orgFinder.On("FindByName", mock.Anything, fixtureGuestOrgName).Return(org, nil)
	return orgFinder
}

func newGuestAuthUserFinder(t *testing.T, enabled bool) *MockAppUserByLoginIDFinder {
	t.Helper()
	user := domainuser.ReconstructAppUser(fixtureAppUserID, fixtureOrgID, fixtureGuestLoginID, "", enabled)
	userFinder := NewMockAppUserByLoginIDFinder(t)
	userFinder.On("FindByLoginID", mock.Anything, fixtureOrgID, fixtureGuestLoginID).Return(user, nil)
	return userFinder
}

func newGuestAuthInput(t *testing.T) *authservice.GuestAuthenticateInput {
	t.Helper()
	input, err := authservice.NewGuestAuthenticateInput(fixtureGuestOrgName)
	require.NoError(t, err)
	return input
}

func Test_GuestAuthenticateQuery_GuestAuthenticate_shouldReturnGuestUser_whenGuestIsEnabled(t *testing.T) {
	t.Parallel()

	// given
	q := authusecase.NewGuestAuthenticateQuery(newGuestAuthOrgFinder(t), newGuestAuthUserFinder(t, true))

	// when
	output, err := q.GuestAuthenticate(context.Background(), newGuestAuthInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, authservice.GuestAuthenticateOutput{UserID: fixtureAppUserID, LoginID: string(fixtureGuestLoginID), OrganizationName: fixtureGuestOrgName}, *output)
}

func Test_GuestAuthenticateQuery_GuestAuthenticate_shouldReturnErrUnauthenticated_whenOrganizationDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	orgFinder := NewMockOrganizationFinder(t)
	orgFinder.On("FindByName", mock.Anything, fixtureGuestOrgName).Return(nil, domain.ErrOrganizationNotFound)
	q := authusecase.NewGuestAuthenticateQuery(orgFinder, NewMockAppUserByLoginIDFinder(t))

	// when
	_, err := q.GuestAuthenticate(context.Background(), newGuestAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_GuestAuthenticateQuery_GuestAuthenticate_shouldReturnErrUnauthenticated_whenGuestDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	userFinder := NewMockAppUserByLoginIDFinder(t)
	userFinder.On("FindByLoginID", mock.Anything, fixtureOrgID, fixtureGuestLoginID).Return(nil, domain.ErrAppUserNotFound)
	q := authusecase.NewGuestAuthenticateQuery(newGuestAuthOrgFinder(t), userFinder)

	// when
	_, err := q.GuestAuthenticate(context.Background(), newGuestAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_GuestAuthenticateQuery_GuestAuthenticate_shouldReturnErrUnauthenticated_whenGuestIsDisabled(t *testing.T) {
	t.Parallel()

	// given
	q := authusecase.NewGuestAuthenticateQuery(newGuestAuthOrgFinder(t), newGuestAuthUserFinder(t, false))

	// when
	_, err := q.GuestAuthenticate(context.Background(), newGuestAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_GuestAuthenticateQuery_GuestAuthenticate_shouldReturnStorageError_whenOrganizationLookupFails(t *testing.T) {
	t.Parallel()

	// given
	orgFinder := NewMockOrganizationFinder(t)
	orgFinder.On("FindByName", mock.Anything, fixtureGuestOrgName).Return(nil, errAuthStorage)
	q := authusecase.NewGuestAuthenticateQuery(orgFinder, NewMockAppUserByLoginIDFinder(t))

	// when
	_, err := q.GuestAuthenticate(context.Background(), newGuestAuthInput(t))

	// then
	require.ErrorIs(t, err, errAuthStorage)
}

func Test_GuestAuthenticateQuery_GuestAuthenticate_shouldReturnStorageError_whenGuestLookupFails(t *testing.T) {
	t.Parallel()

	// given
	userFinder := NewMockAppUserByLoginIDFinder(t)
	userFinder.On("FindByLoginID", mock.Anything, fixtureOrgID, fixtureGuestLoginID).Return(nil, errAuthStorage)
	q := authusecase.NewGuestAuthenticateQuery(newGuestAuthOrgFinder(t), userFinder)

	// when
	_, err := q.GuestAuthenticate(context.Background(), newGuestAuthInput(t))

	// then
	require.ErrorIs(t, err, errAuthStorage)
}
