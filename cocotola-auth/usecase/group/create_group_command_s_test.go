package group_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
	groupservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/group"
	groupusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/group"
)

func newCreateGroupInput(t *testing.T) *groupservice.CreateGroupInput {
	t.Helper()
	input, err := groupservice.NewCreateGroupInput(fixtureOperatorID, fixtureOrgName, fixtureGroupName)
	require.NoError(t, err)
	return input
}

func Test_CreateGroupCommand_CreateGroup_shouldReturnForbidden_whenOperatorIsNotAllowed(t *testing.T) {
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
		fixtureOperatorID,
		domainrbac.ActionCreateGroup(),
		domainrbac.ResourceAny(),
	).Return(false, nil)

	groupRepo := newMockgroupSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := groupusecase.NewCreateGroupCommand(groupRepo, orgRepo, publisher, authChecker)
	input := newCreateGroupInput(t)

	// when
	output, err := cmd.CreateGroup(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Nil(t, output)
}

func Test_CreateGroupCommand_CreateGroup_shouldReturnError_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(nil, domain.ErrOrganizationNotFound)

	authChecker := newMockauthorizationChecker(t)
	groupRepo := newMockgroupSaver(t)
	publisher := newMockeventPublisher(t)

	cmd := groupusecase.NewCreateGroupCommand(groupRepo, orgRepo, publisher, authChecker)
	input := newCreateGroupInput(t)

	// when
	output, err := cmd.CreateGroup(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrOrganizationNotFound)
	require.Nil(t, output)
}

func Test_CreateGroupCommand_CreateGroup_shouldCreateGroupAndPublishEvent_whenAllowed(t *testing.T) {
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
		fixtureOperatorID,
		domainrbac.ActionCreateGroup(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	groupRepo := newMockgroupSaver(t)
	groupRepo.On("Save", mock.Anything, mock.AnythingOfType("*group.Group")).Return(nil)

	var published domain.GroupCreated
	publisher := newMockeventPublisher(t)
	publisher.On("Publish", mock.AnythingOfType("domain.GroupCreated")).
		Run(func(args mock.Arguments) {
			event, ok := args.Get(0).(domain.GroupCreated)
			require.True(t, ok, "event must be domain.GroupCreated")
			published = event
		}).
		Return()

	cmd := groupusecase.NewCreateGroupCommand(groupRepo, orgRepo, publisher, authChecker)
	input := newCreateGroupInput(t)

	// when
	output, err := cmd.CreateGroup(ctx, input)

	// then
	require.NoError(t, err)
	assert.False(t, output.GroupID.IsZero())
	assert.Equal(t, fixtureOrgID, output.OrganizationID)
	assert.Equal(t, fixtureGroupName, output.Name)
	assert.True(t, output.Enabled)
	assert.Equal(t, domain.EventTypeGroupCreated, published.EventType())
	assert.Equal(t, fixtureOrgID, published.OrganizationID())
	assert.Equal(t, fixtureGroupName, published.Name())
	assert.False(t, published.GroupID().IsZero())
}

func Test_CreateGroupCommand_CreateGroup_shouldReturnError_whenGroupRepoSaveFails(t *testing.T) {
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
		fixtureOperatorID,
		domainrbac.ActionCreateGroup(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	saveErr := errors.New("write failed")
	groupRepo := newMockgroupSaver(t)
	groupRepo.On("Save", mock.Anything, mock.AnythingOfType("*group.Group")).Return(saveErr)

	publisher := newMockeventPublisher(t)

	cmd := groupusecase.NewCreateGroupCommand(groupRepo, orgRepo, publisher, authChecker)
	input := newCreateGroupInput(t)

	// when
	output, err := cmd.CreateGroup(ctx, input)

	// then
	require.ErrorIs(t, err, saveErr)
	require.Nil(t, output)
}
