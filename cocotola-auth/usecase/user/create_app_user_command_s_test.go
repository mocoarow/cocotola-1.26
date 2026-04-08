package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainuser "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/user"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
	userusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/user"
)

func Test_CreateAppUserCommand_CreateAppUser_shouldCreateUser_whenOrganizationHasCapacity(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	orgName := "test-org"
	loginID := "user1@example.com"
	password := "securepass"
	hashedPassword := "$2a$10$hashed"
	generatedUserID := 42
	maxActiveUsers := 10

	org := domain.ReconstructOrganization(orgID, orgName, maxActiveUsers, 5)

	idProviderMock := newMockappUserIDProvider(t)
	idProviderMock.On("NextID", mock.Anything).Return(generatedUserID, nil)

	saverMock := newMockappUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.MatchedBy(func(u *domainuser.AppUser) bool {
		return u.ID() == generatedUserID && string(u.LoginID()) == loginID && u.HashedPassword() == hashedPassword
	})).Return(nil)

	orgRepoMock := newMockorganizationFinderByName(t)
	orgRepoMock.On("FindByName", mock.Anything, orgName).Return(org, nil)

	publisherMock := newMockeventPublisher(t)
	publisherMock.On("Publish", mock.Anything)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(idProviderMock, saverMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationName: orgName, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.NoError(t, err)
	require.Equal(t, generatedUserID, output.AppUserID)
	require.Equal(t, orgID, output.OrganizationID)
	require.Equal(t, loginID, output.LoginID)
	require.True(t, output.Enabled)
	publisherMock.AssertCalled(t, "Publish", mock.Anything)
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnForbidden_whenNotAuthorized(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	orgName := "test-org"
	loginID := "user1@example.com"
	password := "securepass"
	maxActiveUsers := 10

	org := domain.ReconstructOrganization(orgID, orgName, maxActiveUsers, 5)

	idProviderMock := newMockappUserIDProvider(t)
	saverMock := newMockappUserSaver(t)

	orgRepoMock := newMockorganizationFinderByName(t)
	orgRepoMock.On("FindByName", mock.Anything, orgName).Return(org, nil)

	publisherMock := newMockeventPublisher(t)
	hasherMock := newMockpasswordHasher(t)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny()).Return(false, nil)

	cmd := userusecase.NewCreateAppUserCommand(idProviderMock, saverMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationName: orgName, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Nil(t, output)
	saverMock.AssertNotCalled(t, "Save")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgName := "nonexistent-org"
	loginID := "user1@example.com"
	password := "securepass"

	idProviderMock := newMockappUserIDProvider(t)
	saverMock := newMockappUserSaver(t)

	orgRepoMock := newMockorganizationFinderByName(t)
	orgRepoMock.On("FindByName", mock.Anything, orgName).Return(nil, domain.ErrOrganizationNotFound)

	publisherMock := newMockeventPublisher(t)
	hasherMock := newMockpasswordHasher(t)
	authCheckerMock := newMockauthorizationChecker(t)

	cmd := userusecase.NewCreateAppUserCommand(idProviderMock, saverMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationName: orgName, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	require.Nil(t, output)
	saverMock.AssertNotCalled(t, "Save")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenCreateAppUserFails(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	orgName := "test-org"
	loginID := "user1@example.com"
	password := "securepass"
	hashedPassword := "$2a$10$hashed"
	maxActiveUsers := 10
	dbErr := errors.New("db connection error")

	org := domain.ReconstructOrganization(orgID, orgName, maxActiveUsers, 5)

	idProviderMock := newMockappUserIDProvider(t)
	idProviderMock.On("NextID", mock.Anything).Return(99, nil)

	saverMock := newMockappUserSaver(t)
	saverMock.On("Save", mock.Anything, mock.Anything).Return(dbErr)

	orgRepoMock := newMockorganizationFinderByName(t)
	orgRepoMock.On("FindByName", mock.Anything, orgName).Return(org, nil)

	publisherMock := newMockeventPublisher(t)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(idProviderMock, saverMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationName: orgName, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, dbErr)
	require.Nil(t, output)
	publisherMock.AssertNotCalled(t, "Publish")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenHasherFails(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	orgName := "test-org"
	loginID := "user1@example.com"
	password := "securepass"
	maxActiveUsers := 10
	hashErr := errors.New("hash failure")

	org := domain.ReconstructOrganization(orgID, orgName, maxActiveUsers, 5)

	idProviderMock := newMockappUserIDProvider(t)
	saverMock := newMockappUserSaver(t)

	orgRepoMock := newMockorganizationFinderByName(t)
	orgRepoMock.On("FindByName", mock.Anything, orgName).Return(org, nil)

	publisherMock := newMockeventPublisher(t)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return("", hashErr)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domainrbac.ActionCreateUser(), domainrbac.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(idProviderMock, saverMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationName: orgName, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, hashErr)
	require.Nil(t, output)
	saverMock.AssertNotCalled(t, "Save")
}
