package sharing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
)

func newListSharedInput(t *testing.T) *referenceservice.ListSharedInput {
	t.Helper()
	input, err := referenceservice.NewListSharedInput(fixtureOperatorID, fixtureOrganizationID)
	require.NoError(t, err)
	return input
}

func Test_ListSharedQuery_shouldMapReferencesToSharedItems_whenFinderReturnsReferences(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	firstAddedAt := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	secondAddedAt := time.Date(2026, 5, 3, 18, 30, 0, 0, time.UTC)
	refs := []domainreference.WorkbookReference{
		*domainreference.ReconstructWorkbookReference(fixtureOperatorID, "wb-shared-1", firstAddedAt),
		*domainreference.ReconstructWorkbookReference(fixtureOperatorID, "wb-shared-2", secondAddedAt),
	}
	finder := newMockreferenceFinder(t)
	finder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(refs, nil)
	q := sharingusecase.NewListSharedQuery(finder)

	// when
	output, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.NoError(t, err)
	expected := []referenceservice.SharedItem{
		{ReferenceID: "wb-shared-1", WorkbookID: "wb-shared-1", AddedAt: firstAddedAt},
		{ReferenceID: "wb-shared-2", WorkbookID: "wb-shared-2", AddedAt: secondAddedAt},
	}
	assert.Equal(t, expected, output.References)
}

func Test_ListSharedQuery_shouldReturnEmptyList_whenFinderReturnsNothing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockreferenceFinder(t)
	finder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(nil, nil)
	q := sharingusecase.NewListSharedQuery(finder)

	// when
	output, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.NoError(t, err)
	assert.Empty(t, output.References)
}

func Test_ListSharedQuery_shouldReturnError_whenFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finderErr := errors.New("firestore unavailable")
	finder := newMockreferenceFinder(t)
	finder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(nil, finderErr)
	q := sharingusecase.NewListSharedQuery(finder)

	// when
	_, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.ErrorIs(t, err, finderErr)
}
