package workbook_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
	workbookusecase "github.com/mocoarow/cocotola-1.26/cocotola-question/usecase/workbook"
)

// workbookResourceMatcher asserts the policy target is a workbook resource
// whose ID is non-empty. The ID itself is generated dynamically by the command
// (UUID v7), so tests can only verify the resource shape rather than the
// exact value.
var workbookResourceMatcher = mock.MatchedBy(func(r domain.Resource) bool {
	const prefix = "workbook:"
	return strings.HasPrefix(r.Value(), prefix) && len(r.Value()) > len(prefix)
})

// expectAllWorkbookPolicies registers a separate mock expectation for each of
// the six (action, workbook-resource) policy grants performed by
// CreateWorkbook. Restoring the per-action expectations keeps the test
// regression-sensitive to silently dropping or replacing one of the granted
// actions.
func expectAllWorkbookPolicies(policyAdder *mockpolicyAdder) {
	actions := []domain.Action{
		domain.ActionViewWorkbook(),
		domain.ActionUpdateWorkbook(),
		domain.ActionDeleteWorkbook(),
		domain.ActionCreateQuestion(),
		domain.ActionUpdateQuestion(),
		domain.ActionDeleteQuestion(),
	}
	for _, action := range actions {
		policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, action, workbookResourceMatcher, domain.EffectAllow()).Return(nil).Once()
	}
}

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

	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb != nil &&
			wb.SpaceID() == fixtureSpaceID &&
			wb.OwnerID() == fixtureOperatorID &&
			wb.OrganizationID() == fixtureOrganizationID &&
			wb.Title() == "Test Workbook" &&
			wb.Description() == "description" &&
			wb.Visibility().Value() == "private" &&
			wb.Language().Value() == "ja" &&
			wb.Version() == 0 &&
			wb.ID() != ""
	})).Return(nil)

	policyAdder := newMockpolicyAdder(t)
	expectAllWorkbookPolicies(policyAdder)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbSaver, listFinder, listSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	input := newCreateWorkbookInput(t)

	// when
	output, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, output.WorkbookID)
	assert.Equal(t, "Test Workbook", output.Title)
	assert.Equal(t, "private", output.Visibility)
	assert.Equal(t, "ja", output.Language)
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

	wbSaver := newMockworkbookSaver(t)
	// Visibility on the saved aggregate must be "public" even though caller sent "private".
	wbSaver.On("Save", mock.Anything, mock.MatchedBy(func(wb *domainworkbook.Workbook) bool {
		return wb != nil && wb.Visibility().Value() == "public"
	})).Return(nil)

	policyAdder := newMockpolicyAdder(t)
	expectAllWorkbookPolicies(policyAdder)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbSaver, listFinder, listSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	// Caller sends visibility=private but PublicSpace must override it to public.
	input, err := workbookservice.NewCreateWorkbookInput(fixtureOperatorID, fixtureOrganizationID, fixtureSpaceID, "Test Workbook", "description", "private", "ja")
	require.NoError(t, err)

	// when
	output, err := cmd.CreateWorkbook(ctx, input)

	// then
	require.NoError(t, err)
	// Visibility on the aggregate (asserted via mock matcher above) and on the
	// output is "public". The input must NOT be mutated.
	assert.Equal(t, "public", output.Visibility)
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

	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	policyErr := errors.New("auth service unavailable")
	policyAdder := newMockpolicyAdder(t)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrganizationID, fixtureOperatorID, domain.ActionViewWorkbook(), workbookResourceMatcher, domain.EffectAllow()).Return(policyErr)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("private", nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbSaver, listFinder, nil, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
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

	wbSaver := newMockworkbookSaver(t)
	wbSaver.On("Save", mock.Anything, mock.Anything).Return(nil)

	policyAdder := newMockpolicyAdder(t)
	expectAllWorkbookPolicies(policyAdder)

	spaceTypeFetcher := newMockspaceTypeFetcher(t)
	spaceTypeFetcher.On("FetchSpaceType", mock.Anything, fixtureSpaceID).Return("private", nil)

	cmd := workbookusecase.NewCreateWorkbookCommand(wbSaver, listFinder, listSaver, maxWbFetcher, spaceTypeFetcher, authChecker, policyAdder)
	input := newCreateWorkbookInput(t)

	// when
	_, err = cmd.CreateWorkbook(ctx, input)

	// then
	require.ErrorIs(t, err, libversioned.ErrConcurrentModification)
}
