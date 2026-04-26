package sharing_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
)

const (
	fixtureOperatorID     = "user-1"
	fixtureOrganizationID = "org-1"
)

func newListPublicInput(t *testing.T, language string) *referenceservice.ListPublicInput {
	t.Helper()
	input, err := referenceservice.NewListPublicInput(fixtureOperatorID, fixtureOrganizationID, language)
	require.NoError(t, err)
	return input
}

func fixturePublicWorkbook(t *testing.T, id string, language domainworkbook.Language) domainworkbook.Workbook {
	t.Helper()
	now := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	wb, err := domainworkbook.NewWorkbook(id, "public-space", "system-owner", fixtureOrganizationID, "Title", "Desc", domainworkbook.VisibilityPublic(), language, now, now)
	require.NoError(t, err)
	return *wb
}

func Test_ListPublicQuery_shouldReturnMatchingLanguageWorkbooks_whenFinderReturnsResults(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockpublicWorkbookFinder(t)
	jaWorkbooks := []domainworkbook.Workbook{
		fixturePublicWorkbook(t, "wb-1", domainworkbook.LanguageJa()),
	}
	finder.On("FindPublicByOrganizationIDAndLanguage", mock.Anything, fixtureOrganizationID, "ja").Return(jaWorkbooks, nil)
	q := sharingusecase.NewListPublicQuery(finder)

	// when
	output, err := q.ListPublic(ctx, newListPublicInput(t, "ja"))

	// then
	require.NoError(t, err)
	require.Len(t, output.Workbooks, 1)
	assert.Equal(t, "wb-1", output.Workbooks[0].WorkbookID)
	assert.Equal(t, "ja", output.Workbooks[0].Language)
}

func Test_ListPublicQuery_shouldReturnEmptyList_whenFinderReturnsNothing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockpublicWorkbookFinder(t)
	finder.On("FindPublicByOrganizationIDAndLanguage", mock.Anything, fixtureOrganizationID, "ja").Return(nil, nil)
	q := sharingusecase.NewListPublicQuery(finder)

	// when
	output, err := q.ListPublic(ctx, newListPublicInput(t, "ja"))

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Workbooks)
}

func Test_ListPublicQuery_shouldPassLanguageToFinder_whenLanguageIsEn(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finder := newMockpublicWorkbookFinder(t)
	enWorkbooks := []domainworkbook.Workbook{
		fixturePublicWorkbook(t, "wb-en-1", domainworkbook.LanguageEn()),
	}
	finder.On("FindPublicByOrganizationIDAndLanguage", mock.Anything, fixtureOrganizationID, "en").Return(enWorkbooks, nil)
	q := sharingusecase.NewListPublicQuery(finder)

	// when
	output, err := q.ListPublic(ctx, newListPublicInput(t, "en"))

	// then
	require.NoError(t, err)
	require.Len(t, output.Workbooks, 1)
	assert.Equal(t, "en", output.Workbooks[0].Language)
}

func Test_ListPublicQuery_shouldReturnError_whenLanguageFailsDomainValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: a value that passes service-layer len=2 but fails the domain
	// regex (uppercase) — the query must reject it before hitting the finder.
	finder := newMockpublicWorkbookFinder(t)
	q := sharingusecase.NewListPublicQuery(finder)
	input := &referenceservice.ListPublicInput{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		Language:       "JA",
	}

	// when
	_, err := q.ListPublic(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
	finder.AssertNotCalled(t, "FindPublicByOrganizationIDAndLanguage")
}

func Test_ListPublicQuery_shouldReturnError_whenFinderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	finderErr := errors.New("firestore unavailable")
	finder := newMockpublicWorkbookFinder(t)
	finder.On("FindPublicByOrganizationIDAndLanguage", mock.Anything, fixtureOrganizationID, "ja").Return(nil, finderErr)
	q := sharingusecase.NewListPublicQuery(finder)

	// when
	_, err := q.ListPublic(ctx, newListPublicInput(t, "ja"))

	// then
	require.ErrorIs(t, err, finderErr)
}
