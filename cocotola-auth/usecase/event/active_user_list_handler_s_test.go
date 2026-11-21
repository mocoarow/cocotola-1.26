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
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/usecase/occ"
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

func newRetryingActiveUserListHandler(t *testing.T) (*eventusecase.ActiveUserListHandler, *mockactiveUserListRepository) {
	t.Helper()
	org := domain.ReconstructOrganization(fixtureOrgID, "test-org", 10, 5)
	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, fixtureOrgID).Return(org, nil)
	activeUserRepoMock := newMockactiveUserListRepository(t)
	activeUserRepoMock.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return(func(context.Context, domain.OrganizationID) *domain.ActiveUserList {
			list, err := domain.NewActiveUserList(fixtureOrgID, nil)
			require.NoError(t, err)
			return list
		}, nil)
	return eventusecase.NewActiveUserListHandler(activeUserRepoMock, orgRepoMock, slog.Default()), activeUserRepoMock
}

func Test_ActiveUserListHandler_Handle_shouldSucceed_whenSaveConflictsOnce(t *testing.T) {
	t.Parallel()

	// given
	handler, activeUserRepoMock := newRetryingActiveUserListHandler(t)
	activeUserRepoMock.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification).Once()
	activeUserRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)
	event := domain.NewAppUserCreated(fixtureAppUserID, fixtureOrgID, "user@example.com", time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.NoError(t, err)
}

func Test_ActiveUserListHandler_Handle_shouldStopRetrying_whenSaveAttemptsAreExhausted(t *testing.T) {
	t.Parallel()

	// given
	handler, activeUserRepoMock := newRetryingActiveUserListHandler(t)
	activeUserRepoMock.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)
	event := domain.NewAppUserCreated(fixtureAppUserID, fixtureOrgID, "user@example.com", time.Now())

	// when
	_ = handler.Handle(context.Background(), event)

	// then
	activeUserRepoMock.AssertNumberOfCalls(t, "Save", occ.MaxAttempts)
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
