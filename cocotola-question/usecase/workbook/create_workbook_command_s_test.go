package workbook_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
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
	input, err := workbookservice.NewCreateWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureSpaceID, "Test Workbook", "description", "private", "ja")
	require.NoError(t, err)
	return input
}

func Test_CreateWorkbookCommand_shouldCreateWorkbook_whenUnderLimit(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, nil)
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	listSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("private", nil)

	wbCreator := newMockworkbookCreator(t)
	wbCreator.On("Create", mock.Anything, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Test Workbook", "description", "private", "ja").Return(fixtureWorkbookID, nil)

	policyAdder := newMockpolicyAdder(t)
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteWorkbook(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateQuestion(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteQuestion(), wbResource, domain.EffectAllow()).Return(nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbCreator, listFinder, listSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	input := newCreateWorkbookInput(t)

	// when
	output, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.NoError(t, err)
	assert.Equal(t, fixtureWorkbookID, output.WorkbookID)
}

func Test_CreateWorkbookCommand_shouldForceVisibilityToPublic_whenSpaceIsPublic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, nil)
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	listSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("public", nil)

	wbCreator := newMockworkbookCreator(t)
	// Visibility passed to Create must be "public" even though caller sent "private".
	wbCreator.On("Create", mock.Anything, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Test Workbook", "description", "public", "ja").Return(fixtureWorkbookID, nil)

	policyAdder := newMockpolicyAdder(t)
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, mock.Anything, wbResource, domain.EffectAllow()).Return(nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbCreator, listFinder, listSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	// Caller sends visibility=private but PublicSpace must override it to public.
	input, err := workbookservice.NewCreateWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureSpaceID, "Test Workbook", "description", "private", "ja")
	require.NoError(t, err)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.NoError(t, err)
	// Visibility passed to workbookRepo.Create was asserted via the mock expectation above
	// (the second-to-last arg is "public"). The input must NOT be mutated.
	assert.Equal(t, "private", input.Visibility)
}

func Test_CreateWorkbookCommand_shouldReturnForbidden_whenNotAllowed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(false, nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(nil, nil, nil, nil, nil, authChecker, nil)
	input := newCreateWorkbookInput(t)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrForbidden)
}

func Test_CreateWorkbookCommand_shouldReturnError_whenAuthCheckFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authErr := errors.New("auth unavailable")
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(false, authErr)

	cmd := workbookusecase.NewCreateWorkbookCommand(nil, nil, nil, nil, nil, authChecker, nil)
	input := newCreateWorkbookInput(t)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, authErr)
}

func Test_CreateWorkbookCommand_shouldReturnLimitReached_whenAtCapacity(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, []string{"wb-a", "wb-b", "wb-c"})
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("private", nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(nil, listFinder, nil, maxWbFetcher, spaceTypeFetcher, authChecker, nil)
	input := newCreateWorkbookInput(t)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, domain.ErrOwnedWorkbookLimitReached)
}

func Test_CreateWorkbookCommand_shouldReturnError_whenPolicyAdderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, nil)
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	wbCreator := newMockworkbookCreator(t)
	wbCreator.On("Create", mock.Anything, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Test Workbook", "description", "private", "ja").Return(fixtureWorkbookID, nil)

	policyErr := errors.New("auth service unavailable")
	policyAdder := newMockpolicyAdder(t)
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource, domain.EffectAllow()).Return(policyErr)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("private", nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbCreator, listFinder, nil, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	input := newCreateWorkbookInput(t)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, policyErr)
}

func Test_CreateWorkbookCommand_shouldReturnError_whenOwnedListSaveFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	spaceResource, err := domain.ResourceSpace(fixtureSpaceID)
	require.NoError(t, err)

	authChecker := newMockauthorizationChecker(t)
	authChecker.On("IsAllowed", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateWorkbook(), spaceResource).Return(true, nil)

	ownedList, _ := domain.NewOwnedWorkbookList(fixtureOperatorID, nil)
	listFinder := newMockownedWorkbookListFinder(t)
	listFinder.On("FindByOwnerID", mock.Anything, fixtureOperatorID).Return(ownedList, nil)

	listSaver := newMockownedWorkbookListSaver(t)
	listSaver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)

	maxWbFetcher := newMockmaxWorkbooksFetcher(t)
	maxWbFetcher.On("FetchMaxWorkbooks", mock.Anything, fixtureOperatorID).Return(3, nil)

	wbCreator := newMockworkbookCreator(t)
	wbCreator.On("Create", mock.Anything, fixtureSpaceID, fixtureOperatorID, fixtureOrganizationID, "Test Workbook", "description", "private", "ja").Return(fixtureWorkbookID, nil)

	policyAdder := newMockpolicyAdder(t)
	wbResource, err := domain.ResourceWorkbook(fixtureWorkbookID)
	require.NoError(t, err)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateWorkbook(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteWorkbook(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionCreateQuestion(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionUpdateQuestion(), wbResource, domain.EffectAllow()).Return(nil)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionDeleteQuestion(), wbResource, domain.EffectAllow()).Return(nil)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("private", nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbCreator, listFinder, listSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	input := newCreateWorkbookInput(t)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}
