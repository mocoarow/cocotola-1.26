package study_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
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
	validateErrorResponse(t, respBytes, "invalid_request", "one of correct, selectedChoiceIds or regions must be provided")
}

func Test_RecordAnswerHandler_shouldReturn400_whenMultipleAnswerFieldsProvided(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
	}{
		{name: "correct and selectedChoiceIds", body: `{"correct":true,"selectedChoiceIds":["c1"]}`},
		{name: "correct and regions", body: `{"correct":true,"regions":[["I","am"]]}`},
		{name: "selectedChoiceIds and regions", body: `{"selectedChoiceIds":["c1"],"regions":[["I","am"]]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			getUsecase := NewMockGetStudyQuestionsUsecase(t)
			recordUsecase := NewMockRecordAnswerUsecase(t)
			r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
			w := httptest.NewRecorder()

			// when
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(tt.body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", "correct, selectedChoiceIds and regions are mutually exclusive")
		})
	}
}

func Test_RecordAnswerHandler_shouldPassRegionsToUsecase_whenOrderingRegionsProvided(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wantRegions := [][]string{{"am", "I"}, {"a", "student"}}
	recordUsecase := NewMockRecordAnswerUsecase(t)
	output := &studyservice.RecordAnswerOutput{
		NextDueAt:          time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
		ConsecutiveCorrect: 1,
		TotalCorrect:       1,
		TotalIncorrect:     0,
	}
	recordUsecase.On("RecordAnswer", mock.Anything, mock.MatchedBy(func(in *studyservice.RecordAnswerInput) bool {
		return in.Regions != nil && assert.ObjectsAreEqual(wantRegions, *in.Regions)
	})).Return(output, nil).Once()

	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()
	body := `{"regions":[["am","I"],["a","student"]]}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_RecordAnswerHandler_shouldReturn400WithGenericMessage_whenRegionsExceedLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		regions [][]string
	}{
		{name: "too many regions", regions: slices.Repeat([][]string{{}}, domainquestion.MaxOrderingSegments+1)},
		{name: "too many chips", regions: [][]string{slices.Repeat([]string{"a"}, domainquestion.MaxOrderingSegments+1)}},
		{name: "too long text", regions: [][]string{{strings.Repeat("a", domainquestion.MaxOrderingTextLength+1), "b"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			getUsecase := NewMockGetStudyQuestionsUsecase(t)
			recordUsecase := NewMockRecordAnswerUsecase(t)
			r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
			w := httptest.NewRecorder()
			body, err := json.Marshal(map[string][][]string{"regions": tt.regions})
			require.NoError(t, err)

			// when
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", http.StatusText(http.StatusBadRequest))
		})
	}
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

func Test_RecordAnswerHandler_shouldReturn400WithGenericMessage_whenSelectedChoiceIDsExceedLimit(t *testing.T) {
	t.Parallel()

	tooMany := make([]string, studyservice.MaxSelectedChoiceIDsCount+1)
	for i := range tooMany {
		tooMany[i] = "c"
	}
	tests := []struct {
		name string
		ids  []string
	}{
		{name: "too many ids", ids: tooMany},
		{name: "too long id", ids: []string{strings.Repeat("c", studyservice.MaxChoiceIDLength+1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			getUsecase := NewMockGetStudyQuestionsUsecase(t)
			recordUsecase := NewMockRecordAnswerUsecase(t)
			r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
			w := httptest.NewRecorder()
			body, err := json.Marshal(map[string][]string{"selectedChoiceIds": tt.ids})
			require.NoError(t, err)

			// when
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/workbook/"+fixtureWorkbookID+"/study/"+fixtureQuestionID+"/answer", bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", http.StatusText(http.StatusBadRequest))
		})
	}
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
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	listUsecase := NewMockListStudyRecordsUsecase(t)
	r := initStudyRouterWithMiddleware(ctx, t, getUsecase, summaryUsecase, recordUsecase, deleteUsecase, listUsecase, noopMiddleware(), fakeOrgResolverMiddleware(fixtureOrganizationID))
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
