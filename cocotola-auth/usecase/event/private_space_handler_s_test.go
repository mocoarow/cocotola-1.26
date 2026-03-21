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

func Test_PrivateSpaceHandler_Handle_shouldCreatePrivateSpace_whenEventValid(t *testing.T) {
	t.Parallel()

	// given
	orgID := 1
	appUserID := 42
	loginID := "user@example.com"
	expectedSpaceID := 10

	spaceRepoMock := newMockspaceCreator(t)
	spaceRepoMock.On("Create", mock.Anything, orgID, appUserID,
		domain.PrivateSpaceKeyName(loginID), "Private(user@example.com)",
		domain.SpaceTypePrivate().Value(), appUserID,
	).Return(expectedSpaceID, nil)

	userSpaceRepoMock := newMockuserSpaceAdder(t)
	userSpaceRepoMock.On("AddUserToSpace", mock.Anything, orgID, appUserID, expectedSpaceID, appUserID).Return(nil)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, userSpaceRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.NoError(t, err)
	spaceRepoMock.AssertCalled(t, "Create", mock.Anything, orgID, appUserID,
		domain.PrivateSpaceKeyName(loginID), "Private(user@example.com)",
		domain.SpaceTypePrivate().Value(), appUserID,
	)
	userSpaceRepoMock.AssertCalled(t, "AddUserToSpace", mock.Anything, orgID, appUserID, expectedSpaceID, appUserID)
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenSpaceCreationFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := 1
	appUserID := 42
	loginID := "user@example.com"
	createErr := errors.New("db error")

	spaceRepoMock := newMockspaceCreator(t)
	spaceRepoMock.On("Create", mock.Anything, orgID, appUserID,
		mock.Anything, mock.Anything, mock.Anything, appUserID,
	).Return(0, createErr)

	userSpaceRepoMock := newMockuserSpaceAdder(t)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, userSpaceRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, createErr)
	userSpaceRepoMock.AssertNotCalled(t, "AddUserToSpace")
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenAddUserToSpaceFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := 1
	appUserID := 42
	loginID := "user@example.com"
	expectedSpaceID := 10
	addErr := errors.New("db error")

	spaceRepoMock := newMockspaceCreator(t)
	spaceRepoMock.On("Create", mock.Anything, orgID, appUserID,
		mock.Anything, mock.Anything, mock.Anything, appUserID,
	).Return(expectedSpaceID, nil)

	userSpaceRepoMock := newMockuserSpaceAdder(t)
	userSpaceRepoMock.On("AddUserToSpace", mock.Anything, orgID, appUserID, expectedSpaceID, appUserID).Return(addErr)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, userSpaceRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, addErr)
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenEventTypeIsWrong(t *testing.T) {
	t.Parallel()

	// given
	spaceRepoMock := newMockspaceCreator(t)
	userSpaceRepoMock := newMockuserSpaceAdder(t)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, userSpaceRepoMock, slog.Default())

	// when
	err := handler.Handle(context.Background(), badEvent{})

	// then
	require.Error(t, err)
	spaceRepoMock.AssertNotCalled(t, "Create")
}

// badEvent is a dummy Event implementation for testing type assertion failure.
type badEvent struct{}

func (badEvent) EventType() string     { return "bad" }
func (badEvent) OccurredAt() time.Time { return time.Now() }
