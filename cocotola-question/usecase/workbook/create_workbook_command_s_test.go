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

const (
	fixtureOperatorID     = "user-1"
	fixtureOrganizationID = "org-1"
	fixtureSpaceID        = "space-1"
	fixtureWorkbookID     = "wb-1"
)

func newCreateWorkbookInput(t *testing.T) *workbookservice.CreateWorkbookInput {
	t.Helper()
	input, err := workbookservice.NewCreateWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureSpaceID, "Test Workbook", "description", "private")
	require.NoError(t, err)
	return input
}

func Test_CreateWorkbookCommand_shouldCreateWorkbook_whenUnderLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny()).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, nil)
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	listSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	wbCreator := newMockworkbookCreator(t)
	wbCreator.On("Create", mock.Anything, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Test Workbook", "description", "private").Return(fixtureWorkbookID, nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbCreator, listFinder, listSaver, maxWbFetcher, authChecker)
	input := newCreateWorkbookInput(t)

	// when
	output, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureWorkbookID, output.WorkbookID)
}

func Test_CreateWorkbookCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny()).Return(false, nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(nil, nil, nil, nil, authChecker)
	input := newCreateWorkbookInput(t)

	// when
	_, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_CreateWorkbookCommand_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny()).Return(false, errors.New("auth unavailable"))

	cmd := workbookusecase.NewCreateWorkbookCommand(nil, nil, nil, nil, authChecker)
	input := newCreateWorkbookInput(t)

	// when
	_, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization check")
}

func Test_CreateWorkbookCommand_shouldReturnLimitReached_whenAtCapacity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny()).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, []string{"wb-a", "wb-b", "wb-c"})
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(nil, listFinder, nil, maxWbFetcher, authChecker)
	input := newCreateWorkbookInput(t)

	// when
	_, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrOwnedWorkbookLimitReached)
}

func Test_CreateWorkbookCommand_shouldReturnError_whenOwnedListSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), domain.ResourceAny()).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, nil)
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	listSaver.On("Save", mock.Anything, mock.Anything).Return(domain.ErrConcurrentModification)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	wbCreator := newMockworkbookCreator(t)
	wbCreator.On("Create", mock.Anything, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Test Workbook", "description", "private").Return(fixtureWorkbookID, nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbCreator, listFinder, listSaver, maxWbFetcher, authChecker)
	input := newCreateWorkbookInput(t)

	// when
	_, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrConcurrentModification)
}
