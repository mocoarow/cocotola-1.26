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

func newUpdateWorkbookInput(t *testing.T, language string) *workbookservice.UpdateWorkbookInput {
	t.Helper()
	input, err := workbookservice.NewUpdateWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, "Updated Title", "updated desc", "private", language)
	require.NoError(t, err)
	return input
}

func newOwnedFixtureWorkbook() *domainworkbook.Workbook {
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Old Title", "old desc", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), time.Now(), time.Now())
}

func Test_UpdateWorkbookCommand_shouldUpdateLanguage_whenOwnerSendsValidLanguage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newOwnedFixtureWorkbook(), nil)

	wbUpdater := newMockworkbookUpdater(t)
	wbUpdater.On("Update", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb.Language().Value() == "en"
	})).Return(nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, wbUpdater, authChecker)

	// when
	output, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.NoError(t, err)
	assert.Equal(t, "en", output.Language)
}

func Test_UpdateWorkbookCommand_shouldReturnError_whenLanguageFailsDomainValidation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: "JA" passes service-level len=2 but the domain language regex rejects uppercase
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newOwnedFixtureWorkbook(), nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, nil, authChecker)
	input := &workbookservice.UpdateWorkbookInput{
		OperatorID:     fixtureOperatorID,
		OrganizationID: fixtureOrganizationID,
		WorkbookID:     fixtureWorkbookID,
		Title:          "Title",
		Description:    "desc",
		Visibility:     "private",
		Language:       "JA",
	}

	// when
	_, err = cmd.UpdateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_UpdateWorkbookCommand_shouldReturnForbidden_whenOperatorIsNotOwner(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource).Return(true, nil)

	otherOwner := domainworkbook.ReconstructWorkbook(fixtureWorkbookID, fixtureSpaceID, "other-user", fixtureOrganizationID, "T", "D", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), time.Now(), time.Now())
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(otherOwner, nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, nil, authChecker)

	// when
	_, err = cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}
