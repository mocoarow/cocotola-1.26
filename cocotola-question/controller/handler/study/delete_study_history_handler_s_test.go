package study_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

func Test_DeleteStudyHistoryHandler_shouldReturn204_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	deleteUsecase.On("DeleteStudyHistory", mock.Anything, mock.Anything).Return(nil).Once()

	r := initStudyRouterWithDelete(ctx, t, deleteUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/workbook/"+fixtureWorkbookID+"/study", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.Bytes())
}

func Test_DeleteStudyHistoryHandler_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	deleteUsecase.On("DeleteStudyHistory", mock.Anything, mock.Anything).Return(domain.ErrForbidden).Once()

	r := initStudyRouterWithDelete(ctx, t, deleteUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/workbook/"+fixtureWorkbookID+"/study", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, respBytes, "forbidden", "Forbidden")
}

func Test_DeleteStudyHistoryHandler_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	deleteUsecase.On("DeleteStudyHistory", mock.Anything, mock.Anything).Return(domain.ErrWorkbookNotFound).Once()

	r := initStudyRouterWithDelete(ctx, t, deleteUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/workbook/"+fixtureWorkbookID+"/study", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "workbook_not_found", "workbook not found")
}

func Test_DeleteStudyHistoryHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: auth middleware does not set UserID
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	r := initStudyRouterWithMiddleware(ctx, t, getUsecase, summaryUsecase, recordUsecase, deleteUsecase, noopMiddleware(), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/workbook/"+fixtureWorkbookID+"/study", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_DeleteStudyHistoryHandler_shouldReturn500_whenUsecaseFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	deleteUsecase.On("DeleteStudyHistory", mock.Anything, mock.Anything).Return(errors.New("unexpected")).Once()

	r := initStudyRouterWithDelete(ctx, t, deleteUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, "/api/v1/workbook/"+fixtureWorkbookID+"/study", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
