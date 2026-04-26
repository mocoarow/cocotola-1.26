package space_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
	spaceservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/space"
	spaceusecase "github.com/mocoarow/cocotola-1.26/cocotola-auth/usecase/space"
)

var (
	fixtureSpaceID = domain.MustParseSpaceID("11111111-1111-7111-8111-111111111111")
	fixtureOrgID   = domain.MustParseOrganizationID("22222222-2222-7222-8222-222222222222")
	fixtureOwnerID = domain.MustParseAppUserID("33333333-3333-7333-8333-333333333333")
)

func fixtureSpace() *domainspace.Space {
	return domainspace.ReconstructSpace(
		fixtureSpaceID,
		fixtureOrgID,
		fixtureOwnerID,
		"test-key",
		"Test Space",
		domainspace.TypePrivate(),
		false,
	)
}

func Test_FindSpaceQuery_shouldReturnSpace_whenSpaceExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := newMockspaceByIDFinder(t)
	repo.On("FindByID", mock.Anything, fixtureSpaceID).Return(fixtureSpace(), nil)

	query := spaceusecase.NewFindSpaceQuery(repo)
	input, err := spaceservice.NewFindSpaceInput(fixtureSpaceID)
	require.NoError(t, err)

	// when
	output, err := query.FindSpace(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureSpaceID, output.Item.SpaceID)
	assert.Equal(t, fixtureOrgID, output.Item.OrganizationID)
	assert.Equal(t, fixtureOwnerID, output.Item.OwnerID)
	assert.Equal(t, "test-key", output.Item.KeyName)
	assert.Equal(t, "Test Space", output.Item.Name)
	assert.Equal(t, "private", output.Item.SpaceType)
	assert.False(t, output.Item.Deleted)
}

func Test_FindSpaceQuery_shouldReturnError_whenSpaceNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repo := newMockspaceByIDFinder(t)
	repo.On("FindByID", mock.Anything, fixtureSpaceID).Return(nil, domain.ErrSpaceNotFound)

	query := spaceusecase.NewFindSpaceQuery(repo)
	input, err := spaceservice.NewFindSpaceInput(fixtureSpaceID)
	require.NoError(t, err)

	// when
	_, err = query.FindSpace(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrSpaceNotFound)
}

func Test_FindSpaceQuery_shouldReturnError_whenRepoFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	repoErr := errors.New("database unavailable")
	repo := newMockspaceByIDFinder(t)
	repo.On("FindByID", mock.Anything, fixtureSpaceID).Return(nil, repoErr)

	query := spaceusecase.NewFindSpaceQuery(repo)
	input, err := spaceservice.NewFindSpaceInput(fixtureSpaceID)
	require.NoError(t, err)

	// when
	_, err = query.FindSpace(ctx, input)

	// then
	require.ErrorIs(t, err, repoErr)
}
