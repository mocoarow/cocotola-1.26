package workbook_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainworkbook "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/workbook"
	workbookservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/workbook"
)

const createWorkbookBody = `{"spaceId":"` + fixtureSpaceID + `","title":"CEFR B1","description":"desc","visibility":"public","language":"ja"}`

func Test_CreateWorkbookHandler_CreateWorkbook_shouldReturn409_whenOwnedWorkbookLimitReached(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createUsecase := NewMockCreateWorkbookUsecase(t)
	createUsecase.On("CreateWorkbook", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("add owned workbook: %w", domain.ErrOwnedWorkbookLimitReached)).Once()
	r := initInternalWorkbookRouter(ctx, t, createUsecase, NewMockUpdateWorkbookUsecase(t))
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

const createWorkbookBodyWithSeedKey = `{"spaceId":"` + fixtureSpaceID + `","title":"CEFR B1","description":"desc","visibility":"public","language":"ja","seedKey":"vocab-v1"}`

func serveCreateWorkbookWithSeedKey(ctx context.Context, t *testing.T, createUsecase *MockCreateWorkbookUsecase) *httptest.ResponseRecorder {
	t.Helper()

	r := initInternalWorkbookRouter(ctx, t, createUsecase, NewMockUpdateWorkbookUsecase(t))
	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/workbook", strings.NewReader(createWorkbookBodyWithSeedKey))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func Test_CreateWorkbookHandler_CreateWorkbook_shouldPassSeedKeyToUsecase_whenRequestHasSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createUsecase := NewMockCreateWorkbookUsecase(t)
	createUsecase.On("CreateWorkbook", mock.Anything, mock.MatchedBy(func(input *workbookservice.CreateWorkbookInput) bool {
		return input.SeedKey == "vocab-v1"
	})).Return(&workbookservice.CreateWorkbookOutput{WorkbookID: fixtureWorkbookID, Visibility: "public"}, nil).Once()

	// when
	w := serveCreateWorkbookWithSeedKey(ctx, t, createUsecase)

	// then
	assert.Equal(t, http.StatusCreated, w.Code)
}

func Test_CreateWorkbookHandler_CreateWorkbook_shouldReturnSeedKey_whenOutputHasSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	createUsecase := NewMockCreateWorkbookUsecase(t)
	createUsecase.On("CreateWorkbook", mock.Anything, mock.Anything).
		Return(&workbookservice.CreateWorkbookOutput{WorkbookID: fixtureWorkbookID, Visibility: "public", SeedKey: "vocab-v1"}, nil).Once()

	// when
	w := serveCreateWorkbookWithSeedKey(ctx, t, createUsecase)

	// then
	require.Equal(t, http.StatusCreated, w.Code)
	var resp api.WorkbookResponse
	require.NoError(t, json.Unmarshal(readBytes(t, w.Body), &resp))
	assert.Equal(t, "vocab-v1", resp.SeedKey)
}

func Test_WorkbookRequest_shouldMirrorDomainMaxSeedKeyLength_whenSeedKeyBindingTagIsRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		request reflect.Type
	}{
		{name: "CreateWorkbookRequest", request: reflect.TypeFor[api.CreateWorkbookRequest]()},
		{name: "UpdateWorkbookRequest", request: reflect.TypeFor[api.UpdateWorkbookRequest]()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			field, ok := tt.request.FieldByName("SeedKey")
			require.True(t, ok)

			// when
			tag := field.Tag.Get("binding")

			// then
			assert.Equal(t, fmt.Sprintf("omitempty,max=%d", domainworkbook.MaxSeedKeyLength), tag)
		})
	}
}
