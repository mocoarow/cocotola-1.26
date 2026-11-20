package study_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func Test_GetStudyQuestionsHandler_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	output := &studyservice.GetStudyQuestionsOutput{
		Questions: []studyservice.QuestionItem{
			{QuestionID: fixtureQuestionID, QuestionType: "word_fill", Content: `{"source":"hello"}`, Tags: []string{"lang:en", "level:2"}, OrderIndex: 3},
			{
				QuestionID: "q-2", QuestionType: "multiple_choice", Content: `{"question":"pick"}`, OrderIndex: 7,
				Audio: &questionservice.AudioOutput{Source: &questionservice.AudioRefOutput{Path: "audio/q-2-src.mp3", DurationSec: 1.5}},
			},
		},
		TotalDue:    5,
		NewCount:    2,
		ReviewCount: 3,
	}
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.Anything).Return(output, nil).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	var resp api.GetStudyQuestionsResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, api.GetStudyQuestionsResponse{
		Questions: []api.StudyQuestionResponse{
			{
				QuestionId: fixtureQuestionID, QuestionType: "word_fill", Content: `{"source":"hello"}`, Tags: []string{"lang:en", "level:2"}, OrderIndex: 3,
				Audio: api.QuestionAudio{Source: api.QuestionAudioRef{}, Target: api.QuestionAudioRef{}},
			},
			{
				QuestionId: "q-2", QuestionType: "multiple_choice", Content: `{"question":"pick"}`, OrderIndex: 7,
				Audio: api.QuestionAudio{Source: api.QuestionAudioRef{Url: "/audio/q-2-src.mp3", DurationSec: 1.5}, Target: api.QuestionAudioRef{}},
			},
		},
		TotalDue:    5,
		NewCount:    2,
		ReviewCount: 3,
	}, resp)
}

func Test_GetStudyQuestionsHandler_shouldPassParsedInputToUsecase_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	output := &studyservice.GetStudyQuestionsOutput{Questions: []studyservice.QuestionItem{}}
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.MatchedBy(func(in *studyservice.GetStudyQuestionsInput) bool {
		return in.OperatorID == fixtureUserID &&
			in.OrganizationID == fixtureOrganizationID &&
			in.WorkbookID == fixtureWorkbookID &&
			in.Limit == 7 &&
			in.Practice
	})).Return(output, nil).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=7&practice=true", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_GetStudyQuestionsHandler_shouldDefaultPracticeToFalse_whenQueryAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	output := &studyservice.GetStudyQuestionsOutput{Questions: []studyservice.QuestionItem{}}
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.MatchedBy(func(in *studyservice.GetStudyQuestionsInput) bool {
		return !in.Practice
	})).Return(output, nil).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_GetStudyQuestionsHandler_shouldReturn400_whenPracticeIsNotBoolean(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10&practice=yes", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "practice must be a boolean")
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

func Test_GetStudyQuestionsHandler_shouldReturn400WithGenericMessage_whenLimitIsOutOfRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		limit string
	}{
		{name: "below min", limit: "0"},
		{name: "above max", limit: "101"},
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
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit="+tt.limit, nil)
			require.NoError(t, err)
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", http.StatusText(http.StatusBadRequest))
		})
	}
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

func Test_GetStudyQuestionsHandler_shouldPassExcludeIDsToUsecase_whenQueryParamRepeated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: repeated excludeIds query params (form/explode style)
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	output := &studyservice.GetStudyQuestionsOutput{
		Questions: []studyservice.QuestionItem{}, TotalDue: 0, NewCount: 0, ReviewCount: 0,
	}
	getUsecase.On("GetStudyQuestions", mock.Anything, mock.MatchedBy(func(in *studyservice.GetStudyQuestionsInput) bool {
		return len(in.ExcludeIDs) == 2 && in.ExcludeIDs[0] == "q-1" && in.ExcludeIDs[1] == "q-2"
	})).Return(output, nil).Once()

	recordUsecase := NewMockRecordAnswerUsecase(t)
	r := initStudyRouter(ctx, t, getUsecase, recordUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10&excludeIds=q-1&excludeIds=q-2", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_GetStudyQuestionsHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
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

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study?limit=10", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", "Unauthorized")
}

func Test_GetStudyQuestionsHandler_shouldReturn401_whenOrganizationIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	getUsecase := NewMockGetStudyQuestionsUsecase(t)
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	recordUsecase := NewMockRecordAnswerUsecase(t)
	deleteUsecase := NewMockDeleteStudyHistoryUsecase(t)
	listUsecase := NewMockListStudyRecordsUsecase(t)
	r := initStudyRouterWithMiddleware(ctx, t, getUsecase, summaryUsecase, recordUsecase, deleteUsecase, listUsecase, fakeAuthMiddleware(fixtureUserID, "org1"), noopMiddleware())
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
