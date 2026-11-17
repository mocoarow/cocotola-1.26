package workbook_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

const listWorkbooksPath = workbookPath + "?spaceId=" + fixtureSpaceID

func Test_ListWorkbooksHandler_ListWorkbooks_shouldReturnWorkbooksInOrder_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	first := fixtureItem("first")
	second := fixtureItem("second")
	second.Visibility = "public"
	usecases := newWorkbookUsecases(t)
	usecases.list.On("ListWorkbooks", mock.Anything, mock.Anything).
		Return(&workbookservice.ListWorkbooksOutput{Workbooks: []workbookservice.Item{first, second}}, nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, listWorkbooksPath)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	var resp api.ListWorkbooksResponse
	require.NoError(t, json.Unmarshal(readBytes(t, w.Body), &resp))
	assert.Equal(t, []api.WorkbookResponse{toWorkbookResponse(first), toWorkbookResponse(second)}, resp.Workbooks)
}

func Test_ListWorkbooksHandler_ListWorkbooks_shouldReturnEmptyArray_whenNoWorkbooks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.list.On("ListWorkbooks", mock.Anything, mock.Anything).
		Return(&workbookservice.ListWorkbooksOutput{Workbooks: nil}, nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, listWorkbooksPath)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"workbooks":[]}`, w.Body.String())
}

func Test_ListWorkbooksHandler_ListWorkbooks_shouldPassOperatorAndSpaceIDToUsecase_whenRequested(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	expected := &workbookservice.ListWorkbooksInput{
		OperatorID:     fixtureUserID,
		OrganizationID: fixtureOrganizationID,
		SpaceID:        fixtureSpaceID,
	}
	usecases := newWorkbookUsecases(t)
	usecases.list.On("ListWorkbooks", mock.Anything, expected).
		Return(&workbookservice.ListWorkbooksOutput{Workbooks: nil}, nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, listWorkbooksPath)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_ListWorkbooksHandler_ListWorkbooks_shouldReturn400_whenSpaceIDIsMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, newWorkbookUsecases(t))

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "invalid_request", "spaceId query parameter is required")
}

func Test_ListWorkbooksHandler_ListWorkbooks_shouldReturn401_whenOperatorIsMissing(t *testing.T) {
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
			w := serve(ctx, t, r, http.MethodGet, listWorkbooksPath)

			// then
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			validateErrorResponse(t, readBytes(t, w.Body), "unauthorized", "Unauthorized")
		})
	}
}

func Test_ListWorkbooksHandler_ListWorkbooks_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.list.On("ListWorkbooks", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("authorize list workbooks: %w", domain.ErrForbidden)).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, listWorkbooksPath)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "forbidden", "Forbidden")
}
