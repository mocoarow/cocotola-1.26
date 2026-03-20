package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	userservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/user"
	userusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/user"
)

func Test_CreateAppUserCommand_CreateAppUser_shouldCreateUser_whenOrganizationHasCapacity(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"
	hashedPassword := "$2a$10$hashed"
	generatedUserID := 42
	maxActiveUsers := 10

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)

	appUserRepoMock := newMockappUserCreator(t)
	appUserRepoMock.On("Create", mock.Anything, orgID, loginID, hashedPassword).Return(generatedUserID, nil)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	publisherMock := newMockeventPublisher(t)
	publisherMock.On("Publish", mock.Anything)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domain.ActionCreateUser(), domain.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.NoError(t, err)
	require.Equal(t, generatedUserID, output.AppUserID)
	require.Equal(t, orgID, output.OrganizationID)
	require.Equal(t, loginID, output.LoginID)
	require.True(t, output.Enabled)
	appUserRepoMock.AssertCalled(t, "Create", mock.Anything, orgID, loginID, hashedPassword)
	publisherMock.AssertCalled(t, "Publish", mock.Anything)
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnForbidden_whenNotAuthorized(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"

	appUserRepoMock := newMockappUserCreator(t)
	orgRepoMock := newMockorganizationFinder(t)
	publisherMock := newMockeventPublisher(t)
	hasherMock := newMockpasswordHasher(t)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domain.ActionCreateUser(), domain.ResourceAny()).Return(false, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Nil(t, output)
	appUserRepoMock.AssertNotCalled(t, "Create")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"

	appUserRepoMock := newMockappUserCreator(t)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(nil, domain.ErrOrganizationNotFound)

	publisherMock := newMockeventPublisher(t)
	hasherMock := newMockpasswordHasher(t)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domain.ActionCreateUser(), domain.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	require.Nil(t, output)
	appUserRepoMock.AssertNotCalled(t, "Create")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenCreateAppUserFails(t *testing.T) {
	t.Parallel()

	// given
	operatorID := 99
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"
	hashedPassword := "$2a$10$hashed"
	maxActiveUsers := 10
	dbErr := errors.New("db connection error")

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)

	appUserRepoMock := newMockappUserCreator(t)
	appUserRepoMock.On("Create", mock.Anything, orgID, loginID, hashedPassword).Return(0, dbErr)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	publisherMock := newMockeventPublisher(t)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domain.ActionCreateUser(), domain.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationID: orgID, LoginID: loginID, Password: password}

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
	loginID := "user1@example.com"
	password := "securepass"
	maxActiveUsers := 10
	hashErr := errors.New("hash failure")

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)

	appUserRepoMock := newMockappUserCreator(t)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	publisherMock := newMockeventPublisher(t)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return("", hashErr)

	authCheckerMock := newMockauthorizationChecker(t)
	authCheckerMock.On("IsAllowed", mock.Anything, orgID, operatorID, domain.ActionCreateUser(), domain.ResourceAny()).Return(true, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, publisherMock, hasherMock, authCheckerMock)
	input := &userservice.CreateAppUserInput{OperatorID: operatorID, OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, hashErr)
	require.Nil(t, output)
	appUserRepoMock.AssertNotCalled(t, "Create")
}
