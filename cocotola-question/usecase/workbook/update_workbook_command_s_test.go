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
	return newUpdateWorkbookInputWithSeedKey(t, language, "")
}

func newUpdateWorkbookInputWithSeedKey(t *testing.T, language string, seedKey string) *workbookservice.UpdateWorkbookInput {
	t.Helper()
	input, err := workbookservice.NewUpdateWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID, "Updated Title", "updated desc", "private", language, seedKey)
	require.NoError(t, err)
	return input
}

// fixtureLoadedTime is the timestamp baked into newOwnedFixtureWorkbook so
// Touch-bumped tests can compare against a known load-time value.
var fixtureLoadedTime = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

func newOwnedFixtureWorkbook() *domainworkbook.Workbook {
	return newOwnedFixtureWorkbookWithSeedKey("")
}

func newOwnedFixtureWorkbookWithSeedKey(seedKey string) *domainworkbook.Workbook {
	return domainworkbook.ReconstructWorkbook(fixtureWorkbookID, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Old Title", "old desc", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), seedKey, 1, fixtureLoadedTime, fixtureLoadedTime)
}

func newAllowedUpdateMocks(t *testing.T, wb *domainworkbook.Workbook) (*mockauthorizationChecker, *mockworkbookFinder) {
	t.Helper()
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(wb, nil)
	return authChecker, wbFinder
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

	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb.Language().Value() == "en"
	})).Return(nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, wbSaver, authChecker)

	// when
	output, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.NoError(t, err)
	assert.Equal(t, "en", output.Language)
}

func Test_UpdateWorkbookCommand_shouldBumpUpdatedAt_whenSavingChanges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newOwnedFixtureWorkbook(), nil)

	beforeUpdate := time.Now()
	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb.UpdatedAt().After(fixtureLoadedTime) && !wb.UpdatedAt().Before(beforeUpdate)
	})).Return(nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, wbSaver, authChecker)

	// when
	output, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.NoError(t, err)
	assert.True(t, output.UpdatedAt.After(fixtureLoadedTime))
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

	otherOwner := domainworkbook.ReconstructWorkbook(fixtureWorkbookID, fixtureSpaceID, "other-user", fixtureOrganizationID, "T", "D", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), "", 1, time.Now(), time.Now())
	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(otherOwner, nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, nil, authChecker)

	// when
	_, err = cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_UpdateWorkbookCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource).Return(false, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(newOwnedFixtureWorkbook(), nil)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, nil, authChecker)

	// when
	_, err = cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_UpdateWorkbookCommand_shouldReturnWorkbookNotFound_whenWorkbookDoesNotExist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(nil, domain.ErrWorkbookNotFound)

	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, nil, authChecker)

	// when
	_, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "en"))

	// then
	require.ErrorIs(t, err, domain.ErrWorkbookNotFound)
}

func Test_UpdateWorkbookCommand_shouldSaveSeedKey_whenWorkbookHasNoSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker, wbFinder := newAllowedUpdateMocks(t, newOwnedFixtureWorkbook())
	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb.SeedKey() == "vocab-v1"
	})).Return(nil)
	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, wbSaver, authChecker)

	// when
	_, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInputWithSeedKey(t, "ja", "vocab-v1"))

	// then
	require.NoError(t, err)
}

func Test_UpdateWorkbookCommand_shouldReturnSeedKey_whenWorkbookHasSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker, wbFinder := newAllowedUpdateMocks(t, newOwnedFixtureWorkbookWithSeedKey("vocab-v1"))
	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.Anything).Return(nil)
	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, wbSaver, authChecker)

	// when
	output, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInput(t, "ja"))

	// then
	require.NoError(t, err)
	assert.Equal(t, "vocab-v1", output.SeedKey)
}

func Test_UpdateWorkbookCommand_shouldKeepSeedKey_whenInputSeedKeyIsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker, wbFinder := newAllowedUpdateMocks(t, newOwnedFixtureWorkbookWithSeedKey("vocab-v1"))
	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb.SeedKey() == "vocab-v1"
	})).Return(nil)
	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, wbSaver, authChecker)

	// when
	_, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInputWithSeedKey(t, "ja", ""))

	// then
	require.NoError(t, err)
}

func Test_UpdateWorkbookCommand_shouldReturnErrInvalidArgument_whenSeedKeyDiffersFromAssignedSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker, wbFinder := newAllowedUpdateMocks(t, newOwnedFixtureWorkbookWithSeedKey("vocab-v1"))
	cmd := workbookusecase.NewUpdateWorkbookCommand(wbFinder, nil, authChecker)

	// when
	_, err := cmd.UpdateWorkbook(ctx, newUpdateWorkbookInputWithSeedKey(t, "ja", "vocab-v2"))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
