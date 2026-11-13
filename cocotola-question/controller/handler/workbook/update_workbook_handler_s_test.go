package workbook_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const updateWorkbookBody = `{"title":"CEFR B1","description":"desc","visibility":"public","language":"ja"}`

func serveUpdateWorkbook(ctx context.Context, t *testing.T, usecaseErr error) *httptest.ResponseRecorder {
	t.Helper()

	updateUsecase := NewMockUpdateWorkbookUsecase(t)
	updateUsecase.On("UpdateWorkbook", mock.Anything, mock.Anything).Return(nil, usecaseErr).Once()
	r := initInternalWorkbookRouter(ctx, t, NewMockCreateWorkbookUsecase(t), updateUsecase)
	w := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, "/api/v1/internal/workbook/wb-1", strings.NewReader(updateWorkbookBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func Test_UpdateWorkbookHandler_UpdateWorkbook_shouldReturn409_whenConcurrentModification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecaseErr := fmt.Errorf("save workbook: %w", libversioned.ErrConcurrentModification)

	// when
	w := serveUpdateWorkbook(ctx, t, usecaseErr)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "conflict", "resource was modified concurrently")
}

func Test_UpdateWorkbookHandler_UpdateWorkbook_shouldReturn400_whenUsecaseRejectsArgument(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecaseErr := fmt.Errorf("new language: %w", domain.ErrInvalidArgument)

	// when
	w := serveUpdateWorkbook(ctx, t, usecaseErr)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, readBytes(t, w.Body), "invalid_request", "Bad Request")
}
