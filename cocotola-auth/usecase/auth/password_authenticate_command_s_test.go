package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
	authusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/auth"
)

const (
	fixtureAuthOrgName  = "password-auth-org"
	fixtureAuthLoginID  = "alice@example.com"
	fixtureAuthPassword = "correct-password"
	fixtureAuthHashed   = "hashed-correct-password"
)

type passwordAuthMocks struct {
	orgFinder   *MockOrganizationFinder
	userFinder  *MockAppUserByLoginIDFinder
	groupFinder *MockGroupFinder
	hasher      *MockPasswordHasher
}

func newPasswordAuthMocks(t *testing.T) passwordAuthMocks {
	t.Helper()
	return passwordAuthMocks{
		orgFinder:   NewMockOrganizationFinder(t),
		userFinder:  NewMockAppUserByLoginIDFinder(t),
		groupFinder: NewMockGroupFinder(t),
		hasher:      NewMockPasswordHasher(t),
	}
}

func (m passwordAuthMocks) query() *authusecase.PasswordAuthenticateQuery {
	return authusecase.NewPasswordAuthenticateQuery(m.orgFinder, m.userFinder, m.groupFinder, m.hasher)
}

func (m passwordAuthMocks) givenOrganization() {
	org := domain.ReconstructOrganization(fixtureOrgID, fixtureAuthOrgName, 100, 50)
	m.orgFinder.On("FindByName", mock.Anything, fixtureAuthOrgName).Return(org, nil)
}

func (m passwordAuthMocks) givenUser(enabled bool) {
	user := domainuser.ReconstructAppUser(fixtureAppUserID, fixtureOrgID, fixtureAuthLoginID, fixtureAuthHashed, enabled)
	m.userFinder.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID(fixtureAuthLoginID)).Return(user, nil)
}

func (m passwordAuthMocks) givenPasswordMatches() {
	m.hasher.On("Compare", fixtureAuthHashed, fixtureAuthPassword).Return(nil)
}

func newPasswordAuthInput(t *testing.T) *authservice.PasswordAuthenticateInput {
	t.Helper()
	input, err := authservice.NewPasswordAuthenticateInput(fixtureAuthLoginID, fixtureAuthPassword, fixtureAuthOrgName)
	require.NoError(t, err)
	return input
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnUser_whenCredentialsAreValid(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.givenOrganization()
	m.givenUser(true)
	m.givenPasswordMatches()
	m.groupFinder.On("GetGroupsForUser", mock.Anything, fixtureOrgID, fixtureAppUserID).Return([]string{"org:members"}, nil)

	// when
	output, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, authservice.PasswordAuthenticateOutput{UserID: fixtureAppUserID, LoginID: fixtureAuthLoginID, OrganizationName: fixtureAuthOrgName}, *output)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnErrUnauthenticated_whenOrganizationDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.orgFinder.On("FindByName", mock.Anything, fixtureAuthOrgName).Return(nil, domain.ErrOrganizationNotFound)

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnErrUnauthenticated_whenUserDoesNotExist(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.givenOrganization()
	m.userFinder.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID(fixtureAuthLoginID)).Return(nil, domain.ErrAppUserNotFound)

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnErrUnauthenticated_whenUserIsDisabled(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.givenOrganization()
	m.givenUser(false)

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnErrUnauthenticated_whenPasswordIsWrong(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.givenOrganization()
	m.givenUser(true)
	m.hasher.On("Compare", fixtureAuthHashed, fixtureAuthPassword).Return(errors.New("mismatch"))

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrUnauthenticated)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnErrUnauthenticated_whenUserIsInLoginDeniedGroup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		group string
	}{
		{name: "system_admin", group: "system_admin"},
		{name: "system_owner", group: domainrbac.GroupSystemOwner().Value()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			m := newPasswordAuthMocks(t)
			m.givenOrganization()
			m.givenUser(true)
			m.givenPasswordMatches()
			m.groupFinder.On("GetGroupsForUser", mock.Anything, fixtureOrgID, fixtureAppUserID).Return([]string{"org:members", tt.group}, nil)

			// when
			_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

			// then
			require.ErrorIs(t, err, domain.ErrUnauthenticated)
		})
	}
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnStorageError_whenOrganizationLookupFails(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.orgFinder.On("FindByName", mock.Anything, fixtureAuthOrgName).Return(nil, errAuthStorage)

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, errAuthStorage)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnStorageError_whenUserLookupFails(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.givenOrganization()
	m.userFinder.On("FindByLoginID", mock.Anything, fixtureOrgID, domain.LoginID(fixtureAuthLoginID)).Return(nil, errAuthStorage)

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, errAuthStorage)
}

func Test_PasswordAuthenticateQuery_PasswordAuthenticate_shouldReturnStorageError_whenGroupLookupFails(t *testing.T) {
	t.Parallel()

	// given
	m := newPasswordAuthMocks(t)
	m.givenOrganization()
	m.givenUser(true)
	m.givenPasswordMatches()
	m.groupFinder.On("GetGroupsForUser", mock.Anything, fixtureOrgID, fixtureAppUserID).Return(nil, errAuthStorage)

	// when
	_, err := m.query().PasswordAuthenticate(context.Background(), newPasswordAuthInput(t))

	// then
	require.ErrorIs(t, err, errAuthStorage)
}
