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

func Test_GetStudySummaryHandler_shouldReturn200_whenValidRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	output := &studyservice.GetStudySummaryOutput{
		NewCount:               5,
		ReviewCount:            12,
		TotalDue:               17,
		ReviewRatioNumerator:   9,
		ReviewRatioDenominator: 10,
	}
	summaryUsecase.On("GetStudySummary", mock.Anything, mock.Anything).Return(output, nil).Once()

	r := initStudyRouterWithSummary(ctx, t, summaryUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/summary", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)

	newCount := parseExpr(t, "$.newCount").Get(jsonObj)
	require.Len(t, newCount, 1)
	assert.EqualValues(t, 5, newCount[0])

	reviewCount := parseExpr(t, "$.reviewCount").Get(jsonObj)
	require.Len(t, reviewCount, 1)
	assert.EqualValues(t, 12, reviewCount[0])

	totalDue := parseExpr(t, "$.totalDue").Get(jsonObj)
	require.Len(t, totalDue, 1)
	assert.EqualValues(t, 17, totalDue[0])

	num := parseExpr(t, "$.reviewRatioNumerator").Get(jsonObj)
	require.Len(t, num, 1)
	assert.EqualValues(t, 9, num[0])

	denom := parseExpr(t, "$.reviewRatioDenominator").Get(jsonObj)
	require.Len(t, denom, 1)
	assert.EqualValues(t, 10, denom[0])
}

func Test_GetStudySummaryHandler_shouldReturn400_whenPracticeIsNotBoolean(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	r := initStudyRouterWithSummary(ctx, t, summaryUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/summary?practice=maybe", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "practice must be a boolean")
}

func Test_GetStudySummaryHandler_shouldReturn403_whenForbidden(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	summaryUsecase.On("GetStudySummary", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("get study summary: %w", domain.ErrForbidden)).Once()

	r := initStudyRouterWithSummary(ctx, t, summaryUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/summary", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
	validateErrorResponse(t, respBytes, "forbidden", "Forbidden")
}

func Test_GetStudySummaryHandler_shouldReturn404_whenWorkbookNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	summaryUsecase := NewMockGetStudySummaryUsecase(t)
	summaryUsecase.On("GetStudySummary", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("find workbook: %w", domain.ErrWorkbookNotFound)).Once()

	r := initStudyRouterWithSummary(ctx, t, summaryUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/"+fixtureWorkbookID+"/study/summary", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "workbook_not_found", "workbook not found")
}
