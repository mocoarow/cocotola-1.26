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
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
	spaceusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/space"
)

var (
	fixturePublicSpaceID  = domain.MustParseSpaceID("00000000-0000-7000-8000-000000000001")
	fixturePrivateSpaceID = domain.MustParseSpaceID("00000000-0000-7000-8000-000000000002")
)

func fixturePublicSpace(t *testing.T) domainspace.Space {
	t.Helper()
	return *domainspace.ReconstructSpace(
		fixturePublicSpaceID,
		fixtureOrgID,
		fixtureOwnerID,
		"public@@acme",
		"Public Space",
		domainspace.TypePublic(),
		false,
	)
}

func fixturePrivateSpace(t *testing.T) domainspace.Space {
	t.Helper()
	return *domainspace.ReconstructSpace(
		fixturePrivateSpaceID,
		fixtureOrgID,
		fixtureOwnerID,
		"private@@owner",
		"Private Space",
		domainspace.TypePrivate(),
		false,
	)
}

func newListSpacesInput(t *testing.T) *spaceservice.ListSpacesInput {
	t.Helper()
	input, err := spaceservice.NewListSpacesInput(fixtureOwnerID, fixtureOrgName)
	require.NoError(t, err)
	return input
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnPublicSpaces_whenOperatorIsAllowed(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return([]domainspace.Space{fixturePublicSpace(t)}, nil)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	require.Len(t, output.Spaces, 1)
	assert.Equal(t, fixturePublicSpaceID, output.Spaces[0].SpaceID)
	assert.Equal(t, "public", output.Spaces[0].SpaceType)
}

func Test_ListSpacesQuery_ListSpaces_shouldIncludePrivateSpace_whenOperatorHasAccess(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(true, nil)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionViewSpace(),
		domainrbac.ResourceSpace(fixturePrivateSpaceID),
	).Return(true, nil)

	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return([]domainspace.Space{fixturePublicSpace(t), fixturePrivateSpace(t)}, nil)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	require.Len(t, output.Spaces, 2)
	assert.Equal(t, fixturePublicSpaceID, output.Spaces[0].SpaceID)
	assert.Equal(t, fixturePrivateSpaceID, output.Spaces[1].SpaceID)
	assert.Equal(t, "private", output.Spaces[1].SpaceType)
}

func Test_ListSpacesQuery_ListSpaces_shouldSkipPrivateSpace_whenOperatorHasNoAccess(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(true, nil)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionViewSpace(),
		domainrbac.ResourceSpace(fixturePrivateSpaceID),
	).Return(false, nil)

	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return([]domainspace.Space{fixturePublicSpace(t), fixturePrivateSpace(t)}, nil)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	require.Len(t, output.Spaces, 1)
	assert.Equal(t, fixturePublicSpaceID, output.Spaces[0].SpaceID)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnEmpty_whenNoSpacesExist(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return([]domainspace.Space{}, nil)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Spaces)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenOrganizationFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repoErr := errors.New("database unavailable")
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(nil, repoErr)

	authChecker := newMockauthorizationChecker(t)
	spaceRepo := newMockspaceFinder(t)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenListAuthCheckerFails(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(false, authErr)

	spaceRepo := newMockspaceFinder(t)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, authErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnForbidden_whenOperatorIsNotAllowedToList(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(false, nil)

	spaceRepo := newMockspaceFinder(t)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenSpaceRepoFails(t *testing.T) {
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
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(true, nil)

	repoErr := errors.New("query failed")
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).Return(nil, repoErr)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenViewSpaceAuthCheckerFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := fixtureOrganization(t)
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(org, nil)

	viewErr := errors.New("view check failed")
	authChecker := newMockauthorizationChecker(t)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionListSpaces(),
		domainrbac.ResourceAny(),
	).Return(true, nil)
	authChecker.On(
		"IsAllowed",
		mock.Anything,
		fixtureOrgID,
		fixtureOwnerID,
		domainrbac.ActionViewSpace(),
		domainrbac.ResourceSpace(fixturePrivateSpaceID),
	).Return(false, viewErr)

	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return([]domainspace.Space{fixturePrivateSpace(t)}, nil)

	query := spaceusecase.NewListSpacesQuery(spaceRepo, orgRepo, authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, viewErr)
}
