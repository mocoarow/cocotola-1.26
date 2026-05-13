package study_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

func Test_ListStudyRecordsHandler_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listUsecase := NewMockListStudyRecordsUsecase(t)
	now := time.Date(2026, 4, 25, 10, 0, 0, 0, time.UTC)
	output := &studyservice.ListStudyRecordsOutput{
		Records: []studyservice.RecordItem{
			{
				WorkbookID:         fixtureWorkbookID,
				QuestionID:         fixtureQuestionID,
				ConsecutiveCorrect: 2,
				LastAnsweredAt:     now,
				NextDueAt:          now.Add(24 * time.Hour),
				TotalCorrect:       3,
				TotalIncorrect:     1,
			},
		},
	}
	listUsecase.On("ListStudyRecords", mock.Anything, mock.Anything).Return(output, nil).Once()

	r := initStudyRouterWithListRecords(ctx, t, listUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/records", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	jsonObj := parseJSON(t, respBytes)

	records := parseExpr(t, "$.records").Get(jsonObj)
	require.Len(t, records, 1)

	qid := parseExpr(t, "$.records[0].questionId").Get(jsonObj)
	require.Len(t, qid, 1)
	assert.Equal(t, fixtureQuestionID, qid[0])

	streak := parseExpr(t, "$.records[0].consecutiveCorrect").Get(jsonObj)
	require.Len(t, streak, 1)
	assert.EqualValues(t, 2, streak[0])
}

func Test_ListStudyRecordsHandler_shouldReturn200WithEmpty_whenNoRecords(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listUsecase := NewMockListStudyRecordsUsecase(t)
	listUsecase.On("ListStudyRecords", mock.Anything, mock.Anything).
		Return(&studyservice.ListStudyRecordsOutput{Records: nil}, nil).Once()

	r := initStudyRouterWithListRecords(ctx, t, listUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/records", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	jsonObj := parseJSON(t, respBytes)
	records := parseExpr(t, "$.records").Get(jsonObj)
	require.Len(t, records, 1)
	assert.Empty(t, records[0])
}

func Test_ListStudyRecordsHandler_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listUsecase := NewMockListStudyRecordsUsecase(t)
	listUsecase.On("ListStudyRecords", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("authorization check: %w", domain.ErrForbidden)).Once()

	r := initStudyRouterWithListRecords(ctx, t, listUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/records", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, respBytes, "forbidden", "Forbidden")
}

func Test_ListStudyRecordsHandler_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listUsecase := NewMockListStudyRecordsUsecase(t)
	listUsecase.On("ListStudyRecords", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound)).Once()

	r := initStudyRouterWithListRecords(ctx, t, listUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/records", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "workbook_not_found", "workbook not found")
}

func Test_ListStudyRecordsHandler_shouldReturn500_whenUsecaseFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	listUsecase := NewMockListStudyRecordsUsecase(t)
	listUsecase.On("ListStudyRecords", mock.Anything, mock.Anything).
		Return(nil, errors.New("unexpected")).Once()

	r := initStudyRouterWithListRecords(ctx, t, listUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/records", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
