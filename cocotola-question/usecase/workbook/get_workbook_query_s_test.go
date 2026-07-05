package workbook_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
	workbookusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/workbook"
)

func newFixturePublicWorkbook() *domainworkbook.Workbook {
	now := time.Now()
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "title", "", domainworkbook.VisibilityPublic(), domainworkbook.LanguageJa(), 1, now, now)
}

func newGetWorkbookInput(t *testing.T) *workbookservice.GetWorkbookInput {
	t.Helper()
	in, err := workbookservice.NewGetWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID)
	require.NoError(t, err)
	return in
}

func Test_GetWorkbookQuery_shouldReturnWorkbook_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newFixturePublicWorkbook(), nil)
	authChecker := newMockauthorizationChecker(t) // no expectations: verifies IsAllowed is not called

	q := workbookusecase.NewGetWorkbookQuery(wbFinder, authChecker)

	// when
	output, err := q.GetWorkbook(ctx, newGetWorkbookInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureWorkbookID, output.WorkbookID)
}

func Test_GetWorkbookQuery_shouldReturnWorkbook_whenPrivateWorkbookAndAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newFixtureWorkbook(fixtureOperatorID), nil)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource).Return(true, nil)

	q := workbookusecase.NewGetWorkbookQuery(wbFinder, authChecker)

	// when
	output, err := q.GetWorkbook(ctx, newGetWorkbookInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureWorkbookID, output.WorkbookID)
}

func Test_GetWorkbookQuery_shouldReturnForbidden_whenPrivateWorkbookAndNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newFixtureWorkbook(fixtureOperatorID), nil)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource).Return(false, nil)

	q := workbookusecase.NewGetWorkbookQuery(wbFinder, authChecker)

	// when
	_, err = q.GetWorkbook(ctx, newGetWorkbookInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_GetWorkbookQuery_shouldReturnError_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	q := workbookusecase.NewGetWorkbookQuery(wbFinder, nil)

	// when
	_, err := q.GetWorkbook(ctx, newGetWorkbookInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}
