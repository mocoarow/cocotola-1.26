package workbook_test

import (
	"context"
	"encoding/json"
	"errors"
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

func Test_GetWorkbookHandler_GetWorkbook_shouldReturnWorkbook_whenFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := fixtureItem("get")
	usecases := newWorkbookUsecases(t)
	usecases.get.On("GetWorkbook", mock.Anything, mock.Anything).
		Return(&workbookservice.GetWorkbookOutput{Item: item}, nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	var resp api.WorkbookResponse
	require.NoError(t, json.Unmarshal(readBytes(t, w.Body), &resp))
	assert.Equal(t, toWorkbookResponse(item), resp)
}

func Test_GetWorkbookHandler_GetWorkbook_shouldPassOperatorAndWorkbookIDToUsecase_whenRequested(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	expected := &workbookservice.GetWorkbookInput{
		OperatorID:     fixtureUserID,
		OrganizationID: fixtureOrganizationID,
		WorkbookID:     fixtureWorkbookID,
	}
	usecases := newWorkbookUsecases(t)
	usecases.get.On("GetWorkbook", mock.Anything, expected).
		Return(&workbookservice.GetWorkbookOutput{Item: fixtureItem("get")}, nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_GetWorkbookHandler_GetWorkbook_shouldReturn401_whenOperatorIsMissing(t *testing.T) {
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
			w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

			// then
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			validateErrorResponse(t, readBytes(t, w.Body), "unauthorized", "Unauthorized")
		})
	}
}

func Test_GetWorkbookHandler_GetWorkbook_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.get.On("GetWorkbook", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound)).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "workbook_not_found", "workbook not found")
}

func Test_GetWorkbookHandler_GetWorkbook_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.get.On("GetWorkbook", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("authorize get workbook: %w", domain.ErrForbidden)).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "forbidden", "Forbidden")
}

func Test_GetWorkbookHandler_GetWorkbook_shouldReturn500_whenUsecaseFailsUnexpectedly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecases := newWorkbookUsecases(t)
	usecases.get.On("GetWorkbook", mock.Anything, mock.Anything).
		Return(nil, errors.New("firestore unavailable")).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "internal_server_error", "Internal Server Error")
}

func Test_GetWorkbookHandler_GetWorkbook_shouldReturnEmptySeedKey_whenWorkbookHasNoSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	item := fixtureItem("get")
	item.SeedKey = ""
	usecases := newWorkbookUsecases(t)
	usecases.get.On("GetWorkbook", mock.Anything, mock.Anything).
		Return(&workbookservice.GetWorkbookOutput{Item: item}, nil).Once()
	r := initWorkbookRouter(ctx, t, fixtureUserID, fixtureOrganizationID, usecases)

	// when
	w := serve(ctx, t, r, http.MethodGet, workbookPath+"/"+fixtureWorkbookID)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(readBytes(t, w.Body), &resp))
	require.Contains(t, resp, "seedKey")
	assert.Empty(t, resp["seedKey"])
}
