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
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	eventusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/event"
)

func Test_PrivateSpaceHandler_Handle_shouldCreatePrivateSpace_whenEventValid(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID
	loginID := "user@example.com"

	var capturedSpaceID domain.SpaceID

	spaceRepoMock := newMockspaceSaver(t)
	spaceRepoMock.On("Save", mock.Anything, mock.MatchedBy(func(s *domainspace.Space) bool {
		return s.OrganizationID() == orgID &&
			s.OwnerID() == appUserID &&
			s.KeyName() == domainspace.PrivateSpaceKeyName(loginID) &&
			s.Name() == "Private(user@example.com)" &&
			s.SpaceType() == domainspace.TypePrivate()
	})).Run(func(args mock.Arguments) {
		s := args.Get(1).(*domainspace.Space)
		capturedSpaceID = s.ID()
	}).Return(nil)

	policyRepoMock := newMockuserPolicyAdder(t)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionViewSpace(), mock.Anything, domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionCreateWorkbook(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionViewWorkbook(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionUpdateWorkbook(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionDeleteWorkbook(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionImportWorkbook(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, policyRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.NoError(t, err)
	require.False(t, capturedSpaceID.IsZero())
	spaceRepoMock.AssertCalled(t, "Save", mock.Anything, mock.Anything)
	policyRepoMock.AssertCalled(t, "AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(capturedSpaceID), domainrbac.EffectAllow(),
	)
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenSpaceCreationFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID
	loginID := "user@example.com"
	saveErr := errors.New("db error")

	spaceRepoMock := newMockspaceSaver(t)
	spaceRepoMock.On("Save", mock.Anything, mock.Anything).Return(saveErr)

	policyRepoMock := newMockuserPolicyAdder(t)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, policyRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, saveErr)
	policyRepoMock.AssertNotCalled(t, "AddPolicyForUser")
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenAddListSpacesPolicyFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID
	loginID := "user@example.com"
	addErr := errors.New("db error")

	spaceRepoMock := newMockspaceSaver(t)
	spaceRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	policyRepoMock := newMockuserPolicyAdder(t)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(addErr)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, policyRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, addErr)
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenAddViewSpacePolicyFails(t *testing.T) {
	t.Parallel()

	// given
	orgID := fixtureOrgID
	appUserID := fixtureAppUserID
	loginID := "user@example.com"
	addErr := errors.New("db error")

	spaceRepoMock := newMockspaceSaver(t)
	spaceRepoMock.On("Save", mock.Anything, mock.Anything).Return(nil)

	policyRepoMock := newMockuserPolicyAdder(t)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), domainrbac.EffectAllow(),
	).Return(nil)
	policyRepoMock.On("AddPolicyForUser", mock.Anything, orgID, appUserID,
		domainrbac.ActionViewSpace(), mock.Anything, domainrbac.EffectAllow(),
	).Return(addErr)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, policyRepoMock, slog.Default())
	event := domain.NewAppUserCreated(appUserID, orgID, loginID, time.Now())

	// when
	err := handler.Handle(context.Background(), event)

	// then
	require.ErrorIs(t, err, addErr)
}

func Test_PrivateSpaceHandler_Handle_shouldReturnError_whenEventTypeIsWrong(t *testing.T) {
	t.Parallel()

	// given
	spaceRepoMock := newMockspaceSaver(t)
	policyRepoMock := newMockuserPolicyAdder(t)

	handler := eventusecase.NewPrivateSpaceHandler(spaceRepoMock, policyRepoMock, slog.Default())

	// when
	err := handler.Handle(context.Background(), badEvent{})

	// then
	require.Error(t, err)
	spaceRepoMock.AssertNotCalled(t, "Save")
}

// badEvent is a dummy Event implementation for testing type assertion failure.
type badEvent struct{}

func (badEvent) EventType() string     { return "bad" }
func (badEvent) OccurredAt() time.Time { return time.Now() }
