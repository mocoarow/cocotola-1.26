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
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
)

var fixtureSharedAddedAt = time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)

func newListSharedInput(t *testing.T) *referenceservice.ListSharedInput {
	t.Helper()
	input, err := referenceservice.NewListSharedInput(fixtureOperatorID, fixtureOrganizationID)
	require.NoError(t, err)
	return input
}

func fixtureReferences(workbookIDs ...string) []domainreference.WorkbookReference {
	refs := make([]domainreference.WorkbookReference, len(workbookIDs))
	for i, id := range workbookIDs {
		refs[i] = *domainreference.ReconstructWorkbookReference(fixtureOperatorID, id, fixtureSharedAddedAt)
	}
	return refs
}

type sharedWorkbookFixture struct {
	id         string
	ownerID    string
	orgID      string
	visibility domainworkbook.Visibility
}

func fixtureSharedWorkbook(t *testing.T, f sharedWorkbookFixture) domainworkbook.Workbook {
	t.Helper()
	now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	wb, err := domainworkbook.NewWorkbook(f.id, "space-other", f.ownerID, f.orgID, "TOEIC 600", "Vocabulary for TOEIC 600", f.visibility, domainworkbook.LanguageEn(), now, now)
	require.NoError(t, err)
	return *wb
}

func publicSharedWorkbook(t *testing.T, id string) domainworkbook.Workbook {
	t.Helper()
	return fixtureSharedWorkbook(t, sharedWorkbookFixture{id: id, ownerID: "other-user", orgID: fixtureOrganizationID, visibility: domainworkbook.VisibilityPublic()})
}

func newListSharedQuery(t *testing.T, refs []domainreference.WorkbookReference, workbooks []domainworkbook.Workbook) *sharingusecase.ListSharedQuery {
	t.Helper()
	refFinder := newMockreferenceFinder(t)
	refFinder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(refs, nil)
	wbFinder := newMockworkbookBatchFinder(t)
	wbFinder.On("FindByIDs", mock.Anything, mock.Anything).Return(workbooks, nil)
	return sharingusecase.NewListSharedQuery(refFinder, wbFinder)
}

func Test_ListSharedQuery_shouldReturnWorkbookDetails_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	q := newListSharedQuery(t, fixtureReferences("wb-1"), []domainworkbook.Workbook{publicSharedWorkbook(t, "wb-1")})

	// when
	output, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.NoError(t, err)
	expected := []referenceservice.SharedItem{{
		ReferenceID: "wb-1",
		WorkbookID:  "wb-1",
		Title:       "TOEIC 600",
		Description: "Vocabulary for TOEIC 600",
		Language:    "en",
		AddedAt:     fixtureSharedAddedAt,
	}}
	assert.Equal(t, expected, output.References)
}

func Test_ListSharedQuery_shouldExcludeWorkbook_whenWorkbookIsNotListable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fixture sharedWorkbookFixture
	}{
		{name: "private", fixture: sharedWorkbookFixture{id: "wb-x", ownerID: "other-user", orgID: fixtureOrganizationID, visibility: domainworkbook.VisibilityPrivate()}},
		{name: "another organization", fixture: sharedWorkbookFixture{id: "wb-x", ownerID: "other-user", orgID: "org-other", visibility: domainworkbook.VisibilityPublic()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			q := newListSharedQuery(t, fixtureReferences("wb-x"), []domainworkbook.Workbook{fixtureSharedWorkbook(t, tt.fixture)})

			// when
			output, err := q.ListShared(ctx, newListSharedInput(t))

			// then
			require.NoError(t, err)
			assert.Empty(t, output.References)
		})
	}
}

func Test_ListSharedQuery_shouldIncludeWorkbook_whenOperatorOwnsIt(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	own := fixtureSharedWorkbook(t, sharedWorkbookFixture{id: "wb-own", ownerID: fixtureOperatorID, orgID: fixtureOrganizationID, visibility: domainworkbook.VisibilityPublic()})
	q := newListSharedQuery(t, fixtureReferences("wb-own"), []domainworkbook.Workbook{own})

	// when
	output, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.NoError(t, err)
	require.Len(t, output.References, 1)
	assert.Equal(t, "wb-own", output.References[0].WorkbookID)
}

func Test_ListSharedQuery_shouldExcludeWorkbook_whenWorkbookDoesNotExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	q := newListSharedQuery(t, fixtureReferences("wb-1", "wb-missing"), []domainworkbook.Workbook{publicSharedWorkbook(t, "wb-1")})

	// when
	output, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.NoError(t, err)
	require.Len(t, output.References, 1)
	assert.Equal(t, "wb-1", output.References[0].WorkbookID)
}

func Test_ListSharedQuery_shouldReturnEmptyList_whenReferenceFinderReturnsNothing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: workbook finder has no expectation, so calling it fails the test
	refFinder := newMockreferenceFinder(t)
	refFinder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(nil, nil)
	q := sharingusecase.NewListSharedQuery(refFinder, newMockworkbookBatchFinder(t))

	// when
	output, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.NoError(t, err)
	assert.Empty(t, output.References)
}

func Test_ListSharedQuery_shouldReturnError_whenReferenceFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finderErr := errors.New("firestore unavailable")
	refFinder := newMockreferenceFinder(t)
	refFinder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(nil, finderErr)
	q := sharingusecase.NewListSharedQuery(refFinder, newMockworkbookBatchFinder(t))

	// when
	_, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.ErrorIs(t, err, finderErr)
}

func Test_ListSharedQuery_shouldReturnError_whenWorkbookFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finderErr := errors.New("firestore unavailable")
	refFinder := newMockreferenceFinder(t)
	refFinder.On("FindByUserID", mock.Anything, fixtureOperatorID).Return(fixtureReferences("wb-1"), nil)
	wbFinder := newMockworkbookBatchFinder(t)
	wbFinder.On("FindByIDs", mock.Anything, []string{"wb-1"}).Return(nil, finderErr)
	q := sharingusecase.NewListSharedQuery(refFinder, wbFinder)

	// when
	_, err := q.ListShared(ctx, newListSharedInput(t))

	// then
	require.ErrorIs(t, err, finderErr)
}
