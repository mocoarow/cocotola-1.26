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
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"
	hashedPassword := "$2a$10$hashed"
	generatedUserID := 42
	maxActiveUsers := 10

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)
	activeUserList, err := domain.NewActiveUserList(orgID, []int{100, 101})
	require.NoError(t, err)

	appUserRepoMock := newMockappUserCreator(t)
	appUserRepoMock.On("Create", mock.Anything, orgID, loginID, hashedPassword).Return(generatedUserID, nil)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeUserRepoMock := newMockactiveUserListRepository(t)
	activeUserRepoMock.On("FindByOrganizationID", mock.Anything, orgID).Return(activeUserList, nil)
	activeUserRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, activeUserRepoMock, hasherMock)
	input := &userservice.CreateAppUserInput{OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.NoError(t, err)
	require.Equal(t, generatedUserID, output.AppUserID)
	appUserRepoMock.AssertCalled(t, "Create", mock.Anything, orgID, loginID, hashedPassword)
	activeUserRepoMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"

	appUserRepoMock := newMockappUserCreator(t)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(nil, domain.ErrOrganizationNotFound)

	activeUserRepoMock := newMockactiveUserListRepository(t)

	hasherMock := newMockpasswordHasher(t)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, activeUserRepoMock, hasherMock)
	input := &userservice.CreateAppUserInput{OrganizationID: orgID, LoginID: loginID, Password: password}

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

	activeUserRepoMock := newMockactiveUserListRepository(t)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, activeUserRepoMock, hasherMock)
	input := &userservice.CreateAppUserInput{OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, dbErr)
	require.Nil(t, output)
	activeUserRepoMock.AssertNotCalled(t, "FindByOrganizationID")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnErrActiveUserLimitReached_whenOrganizationIsFull(t *testing.T) {
	t.Parallel()

	// given
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"
	hashedPassword := "$2a$10$hashed"
	generatedUserID := 42
	maxActiveUsers := 2

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)
	activeUserList, err := domain.NewActiveUserList(orgID, []int{100, 101})
	require.NoError(t, err)

	appUserRepoMock := newMockappUserCreator(t)
	appUserRepoMock.On("Create", mock.Anything, orgID, loginID, hashedPassword).Return(generatedUserID, nil)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeUserRepoMock := newMockactiveUserListRepository(t)
	activeUserRepoMock.On("FindByOrganizationID", mock.Anything, orgID).Return(activeUserList, nil)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return(hashedPassword, nil)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, activeUserRepoMock, hasherMock)
	input := &userservice.CreateAppUserInput{OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, domain.ErrActiveUserLimitReached)
	require.Nil(t, output)
	activeUserRepoMock.AssertNotCalled(t, "Save")
}

func Test_CreateAppUserCommand_CreateAppUser_shouldReturnError_whenHasherFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := 1
	loginID := "user1@example.com"
	password := "securepass"
	maxActiveUsers := 10
	hashErr := errors.New("hash failure")

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)

	appUserRepoMock := newMockappUserCreator(t)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeUserRepoMock := newMockactiveUserListRepository(t)

	hasherMock := newMockpasswordHasher(t)
	hasherMock.On("Hash", password).Return("", hashErr)

	cmd := userusecase.NewCreateAppUserCommand(appUserRepoMock, orgRepoMock, activeUserRepoMock, hasherMock)
	input := &userservice.CreateAppUserInput{OrganizationID: orgID, LoginID: loginID, Password: password}

	// when
	output, err := cmd.CreateAppUser(context.Background(), input)

	// then
	require.ErrorIs(t, err, hashErr)
	require.Nil(t, output)
	appUserRepoMock.AssertNotCalled(t, "Create")
}
