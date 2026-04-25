package study_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func Test_GetStudyQuestionsHandler_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	output := &studyservice.GetStudyQuestionsOutput{
		Questions: []studyservice.QuestionItem{
			{QuestionID: fixtureQuestionID, QuestionType: "word_fill", Content: `{"source":"hello"}`, Tags: []string{"lang:en"}, OrderIndex: 0},
		},
		TotalDue:    1,
		NewCount:    1,
		ReviewCount: 0,
	}
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.Anything).Return(output, nil).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	totalDueExpr := parseExpr(t, "$.totalDue")
	totalDue := totalDueExpr.Get(jsonObj)
	require.Len(t, totalDue, 1)
	assert.EqualValues(t, 1, totalDue[0])

	questionsExpr := parseExpr(t, "$.questions")
	questions := questionsExpr.Get(jsonObj)
	require.Len(t, questions, 1)
}

func Test_GetStudyQuestionsHandler_shouldReturn400_whenLimitMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "limit query parameter is required")
}

func Test_GetStudyQuestionsHandler_shouldReturn400_whenLimitIsNotInteger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=abc", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "limit must be an integer")
}

func Test_GetStudyQuestionsHandler_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("get study questions: %w", domain.ErrForbidden)).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, respBytes, "forbidden", "Forbidden")
}

func Test_GetStudyQuestionsHandler_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound)).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "workbook_not_found", "workbook not found")
}

func Test_GetStudyQuestionsHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouterWithMiddleware(ctx, t, getUsecase, recordUsecase, noopMiddleware(), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", "Unauthorized")
}
