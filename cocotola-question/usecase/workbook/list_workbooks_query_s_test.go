package workbook_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
	workbookusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/workbook"
)

func newListWorkbooksInput(t *testing.T) *workbookservice.ListWorkbooksInput {
	t.Helper()
	in, err := workbookservice.NewListWorkbooksInput(fixtureOperatorID, fixtureOrganizationID, fixtureSpaceID)
	require.NoError(t, err)
	return in
}

func Test_ListWorkbooksQuery_shouldReturnWorkbooks_whenAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), spaceResource).Return(true, nil)
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindBySpaceID", mock.Anything, fixtureSpaceID).Return(nil, nil)

	q := workbookusecase.NewListWorkbooksQuery(wbFinder, authChecker)

	// when
	output, err := q.ListWorkbooks(ctx, newListWorkbooksInput(t))

	// then
	require.NoError(t, err)
	assert.Empty(t, output.Workbooks)
}

func Test_ListWorkbooksQuery_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), spaceResource).Return(false, nil)

	q := workbookusecase.NewListWorkbooksQuery(nil, authChecker)

	// when
	_, err = q.ListWorkbooks(ctx, newListWorkbooksInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ListWorkbooksQuery_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)
	authErr := errors.New("auth unavailable")
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), spaceResource).Return(false, authErr)

	q := workbookusecase.NewListWorkbooksQuery(nil, authChecker)

	// when
	_, err = q.ListWorkbooks(ctx, newListWorkbooksInput(t))

	// then
	require.ErrorIs(t, err, authErr)
}
