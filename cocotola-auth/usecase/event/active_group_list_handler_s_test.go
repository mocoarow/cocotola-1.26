package event_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	eventusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/event"
)

func Test_ActiveGroupListHandler_Handle_shouldAddGroup_whenEventValid(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	groupID := fixtureGroupID
	maxActiveGroups := 10

	org := domain.ReconstructOrganization(orgID, "test-org", 5, maxActiveGroups)
	activeGroupList, err := domain.NewActiveGroupList(orgID, []domain.GroupID{fixtureGroup1, fixtureGroup2})
	require.NoError(t, err)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeGroupRepoMock := newMockactiveGroupListRepository(t)
	activeGroupRepoMock.On("FindByOrganizationID", mock.Anything, orgID).Return(activeGroupList, nil)
	activeGroupRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	handler := eventusecase.NewActiveGroupListHandler(activeGroupRepoMock, orgRepoMock, slog.Default())
	event := domain.NewGroupCreated(groupID, orgID, "test-group", time.Now())

	// when
	err = handler.Handle(context.Background(), event)

	// then
	require.NoError(t, err)
	activeGroupRepoMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
}

func Test_ActiveGroupListHandler_Handle_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	groupID := fixtureGroupID

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(nil, domain.ErrOrganizationNotFound)

	activeGroupRepoMock := newMockactiveGroupListRepository(t)

	handler := eventusecase.NewActiveGroupListHandler(activeGroupRepoMock, orgRepoMock, slog.Default())
	event := domain.NewGroupCreated(groupID, orgID, "test-group", time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	activeGroupRepoMock.AssertNotCalled(t, "FindByOrganizationID")
}

func Test_ActiveGroupListHandler_Handle_shouldReturnError_whenSaveFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	groupID := fixtureGroupID
	maxActiveGroups := 10
	saveErr := errors.New("db error")

	org := domain.ReconstructOrganization(orgID, "test-org", 5, maxActiveGroups)
	activeGroupList, err := domain.NewActiveGroupList(orgID, []domain.GroupID{fixtureGroup1, fixtureGroup2})
	require.NoError(t, err)

	orgRepoMock := newMockorganizationFinder(t)
	orgRepoMock.On("FindByID", mock.Anything, orgID).Return(org, nil)

	activeGroupRepoMock := newMockactiveGroupListRepository(t)
	activeGroupRepoMock.On("FindByOrganizationID", mock.Anything, orgID).Return(activeGroupList, nil)
	activeGroupRepoMock.On("Save", mock.Anything, mock.Anything).Return(saveErr)

	handler := eventusecase.NewActiveGroupListHandler(activeGroupRepoMock, orgRepoMock, slog.Default())
	event := domain.NewGroupCreated(groupID, orgID, "test-group", time.Now())

	// when
	err = handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, saveErr)
}

func Test_ActiveGroupListHandler_Handle_shouldReturnError_whenEventTypeIsWrong(t *testing.T) {
	t.Parallel()

	// given
	orgRepoMock := newMockorganizationFinder(t)
	activeGroupRepoMock := newMockactiveGroupListRepository(t)

	handler := eventusecase.NewActiveGroupListHandler(activeGroupRepoMock, orgRepoMock, slog.Default())

	// when
	err := handler.Handle(context.Background(), badEvent{})

	// then
	require.Error(t, err)
	orgRepoMock.AssertNotCalled(t, "FindByID")
}
