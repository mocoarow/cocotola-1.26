package study_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func Test_RecordAnswerHandler_shouldReturn200_whenCorrectAnswer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	output := &studyservice.RecordAnswerOutput{
		NextDueAt:          time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
		ConsecutiveCorrect: 1,
		TotalCorrect:       1,
		TotalIncorrect:     0,
	}
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(output, nil).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":true}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	consecutiveCorrectExpr := parseExpr(t, "$.consecutiveCorrect")
	consecutiveCorrect := consecutiveCorrectExpr.Get(jsonObj)
	require.Len(t, consecutiveCorrect, 1)
	assert.EqualValues(t, 1, consecutiveCorrect[0])

	totalCorrectExpr := parseExpr(t, "$.totalCorrect")
	totalCorrect := totalCorrectExpr.Get(jsonObj)
	require.Len(t, totalCorrect, 1)
	assert.EqualValues(t, 1, totalCorrect[0])
}

func Test_RecordAnswerHandler_shouldReturn200_whenIncorrectAnswer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	output := &studyservice.RecordAnswerOutput{
		NextDueAt:          time.Date(2026, 4, 25, 10, 10, 0, 0, time.UTC),
		ConsecutiveCorrect: 0,
		TotalCorrect:       0,
		TotalIncorrect:     1,
	}
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(output, nil).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":false}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	totalIncorrectExpr := parseExpr(t, "$.totalIncorrect")
	totalIncorrect := totalIncorrectExpr.Get(jsonObj)
	require.Len(t, totalIncorrect, 1)
	assert.EqualValues(t, 1, totalIncorrect[0])
}

func Test_RecordAnswerHandler_shouldReturn400_whenRequestBodyIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{invalid json}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
}

func Test_RecordAnswerHandler_shouldReturn400_whenBodyHasNeitherField(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "either correct or selectedChoiceIds must be provided")
}

func Test_RecordAnswerHandler_shouldReturn400_whenBothFieldsProvided(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":true,"selectedChoiceIds":["c1"]}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "correct and selectedChoiceIds are mutually exclusive")
}

func Test_RecordAnswerHandler_shouldReturn200_whenMultipleChoiceSelectedChoiceIds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	output := &studyservice.RecordAnswerOutput{
		NextDueAt:          time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
		ConsecutiveCorrect: 1,
		TotalCorrect:       1,
		TotalIncorrect:     0,
	}
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(output, nil).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"selectedChoiceIds":["c1","c2"]}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	totalCorrectExpr := parseExpr(t, "$.totalCorrect")
	totalCorrect := totalCorrectExpr.Get(jsonObj)
	require.Len(t, totalCorrect, 1)
	assert.EqualValues(t, 1, totalCorrect[0])
}

func Test_RecordAnswerHandler_shouldReturn400_whenUsecaseReturnsInvalidArgument(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("evaluate: %w", domain.ErrInvalidArgument)).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":true}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)

	jsonObj := parseJSON(t, respBytes)
	codeExpr := parseExpr(t, "$.code")
	code := codeExpr.Get(jsonObj)
	require.Len(t, code, 1)
	assert.Equal(t, "invalid_request", code[0])
}

func Test_RecordAnswerHandler_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("record answer: %w", domain.ErrForbidden)).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":true}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, respBytes, "forbidden", "Forbidden")
}

func Test_RecordAnswerHandler_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound)).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":true}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "workbook_not_found", "workbook not found")
}

func Test_RecordAnswerHandler_shouldReturn404_whenQuestionNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	recordUsecase := NewMockRecordAnswerUsecase(t)
	recordUsecase.On("RecordAnswer", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("find question: %w", domain.ErrQuestionNotFound)).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"correct":true}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "question_not_found", "question not found")
}

func Test_RecordAnswerHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouterWithMiddleware(ctx, t, getUsecase, recordUsecase, noopMiddleware(), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()
	body := `{"correct":true}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", "Unauthorized")
}
