package event_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	eventusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/event"
)

func Test_ActiveUserListHandler_Handle_shouldAddUser_whenEventValid(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID
	maxActiveUsers := 10

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)
	activeUserList, err := domain.NewActiveUserList(orgID, []domain.AppUserID{fixtureUser1, fixtureUser2})
	require.NoError(t, err)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeUserRepoMock := newMockactiveUserListRepository(t)
	activeUserRepoMock.On("FindByOrganizationID", mock.Anything, orgID).Return(activeUserList, nil)
	activeUserRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	handler := eventusecase.NewActiveUserListHandler(activeUserRepoMock, orgRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, "user@example.com", time.Now())

	// when
	err = handler.Handle(context.Background(), event)

	// then
	require.NoError(t, err)
	activeUserRepoMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func Test_ActiveUserListHandler_Handle_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(nil, domain.ErrOrganizationNotFound)

	activeUserRepoMock := newMockactiveUserListRepository(t)

	handler := eventusecase.NewActiveUserListHandler(activeUserRepoMock, orgRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, "user@example.com", time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	activeUserRepoMock.AssertNotCalled(t, "FindByOrganizationID")
}

func Test_ActiveUserListHandler_Handle_shouldReturnError_whenActiveUserLimitReached(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID
	maxActiveUsers := 2

	org := domain.ReconstructOrganization(orgID, "test-org", maxActiveUsers, 5)
	activeUserList, err := domain.NewActiveUserList(orgID, []domain.AppUserID{fixtureUser1, fixtureUser2})
	require.NoError(t, err)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeUserRepoMock := newMockactiveUserListRepository(t)
	activeUserRepoMock.On("FindByOrganizationID", mock.Anything, orgID).Return(activeUserList, nil)

	handler := eventusecase.NewActiveUserListHandler(activeUserRepoMock, orgRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, "user@example.com", time.Now())

	// when
	err = handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, domain.ErrActiveUserLimitReached)
	activeUserRepoMock.AssertNotCalled(t, "Save")
}
