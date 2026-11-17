package workbook_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

func Test_DeleteWorkbookHandler_DeleteWorkbook_shouldReturn204WithEmptyBody_whenDeleted(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.delete.On("DeleteWorkbook", mock.Anything, mock.Anything).Return(nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodDelete, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func Test_DeleteWorkbookHandler_DeleteWorkbook_shouldPassOperatorAndWorkbookIDToUsecase_whenRequested(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	expected := &workbookservice.DeleteWorkbookInput{
		OperatorID:     fixtureUserID,
		OrganizationID: fixtureOrganizationID,
		WorkbookID:     fixtureWorkbookID,
	}
	usecases := newWorkbookUsecases(t)
	usecases.delete.On("DeleteWorkbook", mock.Anything, expected).Return(nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodDelete, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_DeleteWorkbookHandler_DeleteWorkbook_shouldReturn401_whenOperatorIsMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userID         string
		organizationID string
	}{
		{name: "user ID is missing", userID: "", organizationID: fixtureOrganizationID},
		{name: "organization ID is missing", userID: fixtureUserID, organizationID: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			r := initWorkbookRouter(ctx, t, tt.userID, tt.organizationID, newWorkbookUsecases(t))

			// when
			w := serve(ctx, t, r, http.MethodDelete, workbookPath+"/"+fixtureWorkbookID)

			// then
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			validateErrorResponse(t, readBytes(t, w.Body), "unauthorized", "Unauthorized")
		})
	}
}

func Test_DeleteWorkbookHandler_DeleteWorkbook_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.delete.On("DeleteWorkbook", mock.Anything, mock.Anything).
		Return(fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound)).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodDelete, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "workbook_not_found", "workbook not found")
}

func Test_DeleteWorkbookHandler_DeleteWorkbook_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.delete.On("DeleteWorkbook", mock.Anything, mock.Anything).
		Return(fmt.Errorf("authorize delete workbook: %w", domain.ErrForbidden)).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodDelete, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "forbidden", "Forbidden")
}
