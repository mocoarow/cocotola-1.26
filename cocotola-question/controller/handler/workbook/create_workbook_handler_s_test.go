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

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const createWorkbookBody = `{"spaceId":"space-1","title":"CEFR B1","description":"desc","visibility":"public","language":"ja"}`

func Test_CreateWorkbookHandler_CreateWorkbook_shouldReturn409_whenOwnedWorkbookLimitReached(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createUsecase := NewMockCreateWorkbookUsecase(t)
	createUsecase.On("CreateWorkbook", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("add owned workbook: %w", domain.ErrOwnedWorkbookLimitReached)).Once()
	r := initInternalWorkbookRouter(ctx, t, createUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/workbook", strings.NewReader(createWorkbookBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorResponse(t, respBytes, "owned_workbook_limit_reached", "owned workbook limit reached")
}
