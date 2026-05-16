package space_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
	spaceusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/space"
)

func newCreateSpaceInput(t *testing.T, spaceType string) *spaceservice.CreateSpaceInput {
	t.Helper()
	input, err := spaceservice.NewCreateSpaceInput(fixtureOwnerID, fixtureOrgName, fixtureSpaceName, spaceType)
	require.NoError(t, err)
	return input
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnOutput_whenSpaceTypeIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	spaceRepo := newMockspaceSaver(t)
	spaceRepo.On("Save", mock.Anything, mock.AnythingOfType("*space.Space")).Return(nil)

	publisher := newMockeventPublisher(t)
	publisher.On("Publish", mock.AnythingOfType("domain.SpaceCreated")).Return()

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "public")

	// when
	output, err := cmd.CreateSpace(ctx, input)

	// then
	require.NoError(t, err)
	assert.False(t, output.SpaceID.IsZero())
	assert.Equal(t, fixtureOrgID, output.OrganizationID)
	assert.Equal(t, fixtureOwnerID, output.OwnerID)
	assert.Equal(t, fixturePublicSpaceKey, output.KeyName)
	assert.Equal(t, fixtureSpaceName, output.Name)
	assert.Equal(t, "public", output.SpaceType)
	assert.False(t, output.Deleted)
}

func Test_CreateSpaceCommand_CreateSpace_shouldPublishSpaceCreatedEvent_whenSpaceTypeIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	spaceRepo := newMockspaceSaver(t)
	spaceRepo.On("Save", mock.Anything, mock.AnythingOfType("*space.Space")).Return(nil)

	var published domain.SpaceCreated
	publisher := newMockeventPublisher(t)
	publisher.On("Publish", mock.AnythingOfType("domain.SpaceCreated")).
		Run(func(args mock.Arguments) {
			event, ok := args.Get(0).(domain.SpaceCreated)
			require.True(t, ok, "event must be domain.SpaceCreated")
			published = event
		}).
		Return()

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "public")

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, domain.EventTypeSpaceCreated, published.EventType())
	assert.Equal(t, fixtureOrgID, published.OrganizationID())
	assert.Equal(t, fixtureOwnerID, published.OwnerID())
	assert.Equal(t, fixturePublicSpaceKey, published.KeyName())
	assert.Equal(t, fixtureSpaceName, published.Name())
	assert.Equal(t, "public", published.SpaceType())
	assert.False(t, published.SpaceID().IsZero())
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnError_whenSpaceTypeIsPrivate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	spaceRepo := newMockspaceSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "private")

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.ErrorIs(t, err, spaceusecase.ErrPrivateSpaceMustBeCreatedViaEvent)
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnError_whenOrganizationFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repoErr := errors.New("database unavailable")
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(nil, repoErr)

	authChecker := newMockauthorizationChecker(t)
	spaceRepo := newMockspaceSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "public")

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnError_whenAuthCheckerFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authErr := errors.New("auth backend down")
	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(false, authErr)

	spaceRepo := newMockspaceSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "public")

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnForbidden_whenOperatorIsNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(false, nil)

	spaceRepo := newMockspaceSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "public")

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnError_whenSpaceTypeIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	spaceRepo := newMockspaceSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	// CreateSpaceInput's validator only accepts public/private, so bypass it to reach NewType.
	input := &spaceservice.CreateSpaceInput{
		OperatorID:       fixtureOwnerID,
		OrganizationName: fixtureOrgName,
		Name:             fixtureSpaceName,
		SpaceType:        "invalid",
	}

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_CreateSpaceCommand_CreateSpace_shouldReturnError_whenSpaceRepoSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionCreateSpace(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	saveErr := errors.New("write failed")
	spaceRepo := newMockspaceSaver(t)
	spaceRepo.On("Save", mock.Anything, mock.AnythingOfType("*space.Space")).Return(saveErr)

	publisher := newMockeventPublisher(t)

	cmd := spaceusecase.NewCreateSpaceCommand(spaceRepo, orgRepo, publisher, authChecker)
	input := newCreateSpaceInput(t, "public")

	// when
	_, err := cmd.CreateSpace(ctx, input)

	// then
	require.ErrorIs(t, err, saveErr)
}
