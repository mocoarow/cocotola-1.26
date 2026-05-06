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
	domainreference "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/reference"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
	sharingusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/sharing"
)

const fixtureWorkbookID = "wb-1"

func newShareWorkbookInput(t *testing.T) *referenceservice.ShareWorkbookInput {
	t.Helper()
	input, err := referenceservice.NewShareWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureWorkbookID)
	require.NoError(t, err)
	return input
}

func fixturePublicWorkbookForShare(t *testing.T) *domainworkbook.Workbook {
	t.Helper()
	now := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	wb, err := domainworkbook.NewWorkbook(fixtureWorkbookID, "public-space", "system-owner", fixtureOrganizationID, "Title", "Desc", domainworkbook.VisibilityPublic(), domainworkbook.LanguageJa(), now, now)
	require.NoError(t, err)
	return wb
}

func Test_ShareWorkbookCommand_shouldSaveReferenceKeyedByWorkbookID_whenWorkbookIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionImportWorkbook(), domain.ResourceAny()).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixturePublicWorkbookForShare(t), nil)

	refSaver := newMockreferenceSaver(t)
	refSaver.On("Save", mock.Anything, mock.MatchedBy(func(ref *domainreference.WorkbookReference) bool {
		return ref != nil &&
			ref.UserID() == fixtureOperatorID &&
			ref.WorkbookID() == fixtureWorkbookID &&
			ref.ID() == fixtureWorkbookID
	})).Return(nil)

	cmd := sharingusecase.NewShareWorkbookCommand(refSaver, wbFinder, authChecker)

	// when
	output, err := cmd.ShareWorkbook(ctx, newShareWorkbookInput(t))

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureWorkbookID, output.ReferenceID)
	assert.Equal(t, fixtureWorkbookID, output.WorkbookID)
}

func Test_ShareWorkbookCommand_shouldReturnDuplicateError_whenSaverDetectsExistingReference(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionImportWorkbook(), domain.ResourceAny()).Return(true, nil)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(fixturePublicWorkbookForShare(t), nil)

	refSaver := newMockreferenceSaver(t)
	refSaver.On("Save", mock.Anything, mock.Anything).Return(domain.ErrDuplicateReference)

	cmd := sharingusecase.NewShareWorkbookCommand(refSaver, wbFinder, authChecker)

	// when
	_, err := cmd.ShareWorkbook(ctx, newShareWorkbookInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrDuplicateReference)
}

func Test_ShareWorkbookCommand_shouldReturnForbidden_whenWorkbookIsPrivate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionImportWorkbook(), domain.ResourceAny()).Return(true, nil)

	now := time.Date(2026, 4, 27, 0, 0, 0, 0, time.UTC)
	privateWB, err := domainworkbook.NewWorkbook(fixtureWorkbookID, "space-1", "owner", fixtureOrganizationID, "T", "D", domainworkbook.VisibilityPrivate(), domainworkbook.LanguageJa(), now, now)
	require.NoError(t, err)

	wbFinder := newMockworkbookFinder(t)
	wbFinder.On("FindByID", mock.Anything, fixtureWorkbookID).Return(privateWB, nil)

	cmd := sharingusecase.NewShareWorkbookCommand(nil, wbFinder, authChecker)

	// when
	_, err = cmd.ShareWorkbook(ctx, newShareWorkbookInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ShareWorkbookCommand_shouldReturnForbidden_whenAuthorizationDenies(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionImportWorkbook(), domain.ResourceAny()).Return(false, nil)

	cmd := sharingusecase.NewShareWorkbookCommand(nil, nil, authChecker)

	// when
	_, err := cmd.ShareWorkbook(ctx, newShareWorkbookInput(t))

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_ShareWorkbookCommand_shouldReturnError_whenAuthorizationCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authErr := errors.New("auth unavailable")
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionImportWorkbook(), domain.ResourceAny()).Return(false, authErr)

	cmd := sharingusecase.NewShareWorkbookCommand(nil, nil, authChecker)

	// when
	_, err := cmd.ShareWorkbook(ctx, newShareWorkbookInput(t))

	// then
	require.ErrorIs(t, err, authErr)
}
