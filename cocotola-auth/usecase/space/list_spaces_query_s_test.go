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
	fixturePublicSpaceID       = domain.MustParseSpaceID("00000000-0000-7000-8000-000000000001")
	fixturePrivateSpaceID      = domain.MustParseSpaceID("00000000-0000-7000-8000-000000000002")
	fixtureOtherPrivateSpaceID = domain.MustParseSpaceID("00000000-0000-7000-8000-000000000003")
	fixtureOtherUserID         = domain.MustParseAppUserID("44444444-4444-7444-8444-444444444444")
)

func fixturePublicSpace(t *testing.T) domainspace.Space {
	t.Helper()
	return *domainspace.ReconstructSpace(
		fixturePublicSpaceID,
		fixtureOrgID,
		fixtureOtherUserID,
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

func fixtureOtherPrivateSpace(t *testing.T) domainspace.Space {
	t.Helper()
	return *domainspace.ReconstructSpace(
		fixtureOtherPrivateSpaceID,
		fixtureOrgID,
		fixtureOtherUserID,
		"private@@other",
		"Other Private Space",
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

func newOrgFinderReturningOrg(t *testing.T) *mockorganizationFinderByName {
	t.Helper()
	orgRepo := newMockorganizationFinderByName(t)
	orgRepo.On("FindByName", mock.Anything, fixtureOrgName).Return(fixtureOrganization(t), nil)
	return orgRepo
}

func onIsAllowed(authChecker *mockauthorizationChecker, action domainrbac.Action, resource domainrbac.Resource, allowed bool, err error) {
	authChecker.On("IsAllowed", mock.Anything, fixtureOrgID, fixtureOwnerID, action, resource).Return(allowed, err)
}

func newAuthCheckerAllowingList(t *testing.T, canViewAnySpace bool) *mockauthorizationChecker {
	t.Helper()
	authChecker := newMockauthorizationChecker(t)
	onIsAllowed(authChecker, domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), true, nil)
	onIsAllowed(authChecker, domainrbac.ActionViewSpace(), domainrbac.ResourceAny(), canViewAnySpace, nil)
	return authChecker
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnAllSpacesWithoutPerSpaceCheck_whenOperatorCanViewAnySpace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newAuthCheckerAllowingList(t, true)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).
		Return([]domainspace.Space{fixturePublicSpace(t), fixtureOtherPrivateSpace(t)}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []domain.SpaceID{fixturePublicSpaceID, fixtureOtherPrivateSpaceID}, outputSpaceIDs(output))
}

func Test_ListSpacesQuery_ListSpaces_shouldMapSpaceFields_whenSpaceIsReturned(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newAuthCheckerAllowingList(t, false)
	onIsAllowed(authChecker, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(fixturePrivateSpaceID), true, nil)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).
		Return([]domainspace.Space{fixturePrivateSpace(t)}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []spaceservice.Item{{
		SpaceID:        fixturePrivateSpaceID,
		OrganizationID: fixtureOrgID,
		OwnerID:        fixtureOwnerID,
		KeyName:        "private@@owner",
		Name:           "Private Space",
		SpaceType:      "private",
		Deleted:        false,
	}}, output.Spaces)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnPublicSpaces_whenOperatorCannotViewAnySpace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newAuthCheckerAllowingList(t, false)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).
		Return([]domainspace.Space{fixturePublicSpace(t)}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []domain.SpaceID{fixturePublicSpaceID}, outputSpaceIDs(output))
}

func Test_ListSpacesQuery_ListSpaces_shouldIncludeOwnedPrivateSpace_whenOperatorCanViewIt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newAuthCheckerAllowingList(t, false)
	onIsAllowed(authChecker, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(fixturePrivateSpaceID), true, nil)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).
		Return([]domainspace.Space{fixturePublicSpace(t), fixturePrivateSpace(t)}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []domain.SpaceID{fixturePublicSpaceID, fixturePrivateSpaceID}, outputSpaceIDs(output))
}

func Test_ListSpacesQuery_ListSpaces_shouldSkipOwnedPrivateSpace_whenOperatorCannotViewIt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newAuthCheckerAllowingList(t, false)
	onIsAllowed(authChecker, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(fixturePrivateSpaceID), false, nil)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).
		Return([]domainspace.Space{fixturePublicSpace(t), fixturePrivateSpace(t)}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

	// when
	output, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, []domain.SpaceID{fixturePublicSpaceID}, outputSpaceIDs(output))
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnEmpty_whenNoSpacesExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newAuthCheckerAllowingList(t, false)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).
		Return([]domainspace.Space{}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

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
	query := spaceusecase.NewListSpacesQuery(newMockspaceFinder(t), orgRepo, newMockauthorizationChecker(t))

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenListAuthCheckerFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authErr := errors.New("auth backend down")
	authChecker := newMockauthorizationChecker(t)
	onIsAllowed(authChecker, domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), false, authErr)
	query := spaceusecase.NewListSpacesQuery(newMockspaceFinder(t), newOrgFinderReturningOrg(t), authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, authErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnForbidden_whenOperatorIsNotAllowedToList(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	onIsAllowed(authChecker, domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), false, nil)
	query := spaceusecase.NewListSpacesQuery(newMockspaceFinder(t), newOrgFinderReturningOrg(t), authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenViewAnySpaceAuthCheckerFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	viewErr := errors.New("view any check failed")
	authChecker := newMockauthorizationChecker(t)
	onIsAllowed(authChecker, domainrbac.ActionListSpaces(), domainrbac.ResourceAny(), true, nil)
	onIsAllowed(authChecker, domainrbac.ActionViewSpace(), domainrbac.ResourceAny(), false, viewErr)
	query := spaceusecase.NewListSpacesQuery(newMockspaceFinder(t), newOrgFinderReturningOrg(t), authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, viewErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenFindAllSpacesFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repoErr := errors.New("query failed")
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindByOrganizationID", mock.Anything, fixtureOrgID).Return(nil, repoErr)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), newAuthCheckerAllowingList(t, true))

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenFindPublicOrOwnedSpacesFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repoErr := errors.New("query failed")
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).Return(nil, repoErr)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), newAuthCheckerAllowingList(t, false))

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, repoErr)
}

func Test_ListSpacesQuery_ListSpaces_shouldReturnError_whenViewSpaceAuthCheckerFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	viewErr := errors.New("view check failed")
	authChecker := newAuthCheckerAllowingList(t, false)
	onIsAllowed(authChecker, domainrbac.ActionViewSpace(), domainrbac.ResourceSpace(fixturePrivateSpaceID), false, viewErr)
	spaceRepo := newMockspaceFinder(t)
	spaceRepo.On("FindPublicOrOwnedByOrganizationID", mock.Anything, fixtureOrgID, fixtureOwnerID).
		Return([]domainspace.Space{fixturePrivateSpace(t)}, nil)
	query := spaceusecase.NewListSpacesQuery(spaceRepo, newOrgFinderReturningOrg(t), authChecker)

	// when
	_, err := query.ListSpaces(ctx, newListSpacesInput(t))

	// then
	require.ErrorIs(t, err, viewErr)
}

func outputSpaceIDs(output *spaceservice.ListSpacesOutput) []domain.SpaceID {
	ids := make([]domain.SpaceID, len(output.Spaces))
	for i := range output.Spaces {
		ids[i] = output.Spaces[i].SpaceID
	}
	return ids
}
