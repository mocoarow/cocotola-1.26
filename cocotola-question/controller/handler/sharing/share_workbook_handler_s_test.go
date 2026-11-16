package sharing_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	referenceservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/reference"
)

func Test_ShareWorkbookHandler_ShareWorkbook_shouldReturn201WithReference_whenUsecaseSucceeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	addedAt := time.Date(2026, 10, 2, 8, 30, 0, 0, time.UTC)
	usecases := newSharingUsecases(t)
	usecases.share.On("ShareWorkbook", mock.Anything, &referenceservice.ShareWorkbookInput{
		OperatorID:     fixtureUserID,
		OrganizationID: fixtureOrganizationID,
		WorkbookID:     "wb-42",
	}).Return(&referenceservice.ShareWorkbookOutput{
		ReferenceID: "ref-7",
		WorkbookID:  "wb-42",
		AddedAt:     addedAt,
	}, nil).Once()
	r := initSharingRouter(ctx, t, usecases, fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID))

	// when
	w := serve(ctx, t, r, http.MethodPost, "/api/v1/workbook/wb-42/share")

	// then
	require.Equal(t, http.StatusCreated, w.Code)
	var resp api.ShareWorkbookResponse
	require.NoError(t, json.Unmarshal(readBytes(t, w.Body), &resp))
	assert.Equal(t, api.ShareWorkbookResponse{ReferenceID: "ref-7", WorkbookID: "wb-42", AddedAt: addedAt}, resp)
}

func Test_ShareWorkbookHandler_ShareWorkbook_shouldReturn401_whenOperatorIsMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		authMiddleware gin.HandlerFunc
	}{
		{name: "missing user ID", authMiddleware: fakeAuthMiddleware("", fixtureOrganizationID)},
		{name: "missing organization ID", authMiddleware: fakeAuthMiddleware(fixtureUserID, "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			r := initSharingRouter(ctx, t, newSharingUsecases(t), tt.authMiddleware)

			// when
			w := serve(ctx, t, r, http.MethodPost, "/api/v1/workbook/wb-42/share")

			// then
			require.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Equal(t, api.ErrorResponse{Code: "unauthorized", Message: "Unauthorized"}, readErrorResponse(t, w))
		})
	}
}

func Test_ShareWorkbookHandler_ShareWorkbook_shouldReturn400_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := initSharingRouter(ctx, t, newSharingUsecases(t), fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID))

	// when
	w := serve(ctx, t, r, http.MethodPost, "/api/v1/workbook//share")

	// then
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, api.ErrorResponse{Code: "invalid_request", Message: "workbook ID is required"}, readErrorResponse(t, w))
}

func Test_ShareWorkbookHandler_ShareWorkbook_shouldReturnMappedError_whenUsecaseFails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		usecaseErr   error
		wantStatus   int
		wantResponse api.ErrorResponse
	}{
		{name: "not allowed to import", usecaseErr: domain.ErrForbidden, wantStatus: http.StatusForbidden,
			wantResponse: api.ErrorResponse{Code: "forbidden", Message: "Forbidden"}},
		{name: "workbook not found", usecaseErr: fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound), wantStatus: http.StatusNotFound,
			wantResponse: api.ErrorResponse{Code: "workbook_not_found", Message: "workbook not found"}},
		{name: "already shared", usecaseErr: fmt.Errorf("save reference: %w", domain.ErrDuplicateReference), wantStatus: http.StatusConflict,
			wantResponse: api.ErrorResponse{Code: "duplicate_reference", Message: "workbook reference already exists"}},
		{name: "unexpected error", usecaseErr: errors.New("firestore unavailable"), wantStatus: http.StatusInternalServerError,
			wantResponse: api.ErrorResponse{Code: "internal_server_error", Message: "Internal Server Error"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			usecases := newSharingUsecases(t)
			usecases.share.On("ShareWorkbook", mock.Anything, mock.Anything).Return(nil, tt.usecaseErr).Once()
			r := initSharingRouter(ctx, t, usecases, fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID))

			// when
			w := serve(ctx, t, r, http.MethodPost, "/api/v1/workbook/wb-42/share")

			// then
			require.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantResponse, readErrorResponse(t, w))
		})
	}
}
