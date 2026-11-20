package study_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	studyservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/study"
)

const dashboardURL = "/api/v1/study/dashboard"

func newDashboardRequest(ctx context.Context, t *testing.T, target string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	require.NoError(t, err)
	req.Header.Set("X-Local-Date", "2026-06-14")
	req.Header.Set("X-Local-Timezone", "Asia/Tokyo")
	return req
}

func Test_GetDashboardHandler_shouldReturn200_whenValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockGetDashboardUsecase(t)
	usecase.On("GetDashboard", mock.Anything, mock.MatchedBy(func(in *studyservice.GetDashboardInput) bool {
		return in.OperatorID == fixtureUserID && in.OrganizationID == fixtureOrganizationID && in.Days == 7 && in.TodayDateKey == "2026-06-14"
	})).Return(&studyservice.GetDashboardOutput{
		From: "2026-06-08",
		To:   "2026-06-14",
		Days: []studyservice.DashboardDailyItem{
			{Date: "2026-06-13", AnsweredCount: 4, CorrectCount: 1},
			{Date: "2026-06-14", AnsweredCount: 8, CorrectCount: 6},
		},
		CurrentStreak: 2,
		LongestStreak: 5,
		TodayCount:    8,
		TodayCorrect:  6,
		ActiveDays:    3,
		TotalAnswered: 12,
		TotalCorrect:  7,
	}, nil)

	r := initDashboardRouter(ctx, t, usecase, fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req := newDashboardRequest(ctx, t, dashboardURL+"?days=7")
	r.ServeHTTP(w, req)

	// then
	require.Equal(t, http.StatusOK, w.Code)
	var resp api.GetDashboardResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, api.GetDashboardResponse{
		From: "2026-06-08",
		To:   "2026-06-14",
		Days: []api.DashboardDailyItem{
			{Date: "2026-06-13", AnsweredCount: 4, CorrectCount: 1},
			{Date: "2026-06-14", AnsweredCount: 8, CorrectCount: 6},
		},
		CurrentStreak: 2,
		LongestStreak: 5,
		TodayCount:    8,
		TodayCorrect:  6,
		ActiveDays:    3,
		TotalAnswered: 12,
		TotalCorrect:  7,
	}, resp)
}

func Test_GetDashboardHandler_shouldDefaultDaysTo365_whenQueryAbsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockGetDashboardUsecase(t)
	usecase.On("GetDashboard", mock.Anything, mock.MatchedBy(func(in *studyservice.GetDashboardInput) bool {
		return in.Days == 365
	})).Return(&studyservice.GetDashboardOutput{
		From: "2025-06-15", To: "2026-06-14", Days: []studyservice.DashboardDailyItem{},
	}, nil)

	r := initDashboardRouter(ctx, t, usecase, fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req := newDashboardRequest(ctx, t, dashboardURL)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_GetDashboardHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockGetDashboardUsecase(t)
	r := initDashboardRouter(ctx, t, usecase, noopMiddleware(), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req := newDashboardRequest(ctx, t, dashboardURL)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", http.StatusText(http.StatusUnauthorized))
}

func Test_GetDashboardHandler_shouldReturn401_whenOrganizationIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockGetDashboardUsecase(t)
	r := initDashboardRouter(ctx, t, usecase, fakeAuthMiddleware(fixtureUserID, "org1"), noopMiddleware())
	w := httptest.NewRecorder()

	// when
	req := newDashboardRequest(ctx, t, dashboardURL)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_GetDashboardHandler_shouldReturn400_whenDaysOutOfRange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		days string
	}{
		{name: "tooSmall", days: "3"},
		{name: "tooLarge", days: "9999"},
		{name: "notInteger", days: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			usecase := NewMockGetDashboardUsecase(t)
			r := initDashboardRouter(ctx, t, usecase, fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
			w := httptest.NewRecorder()

			// when
			req := newDashboardRequest(ctx, t, dashboardURL+"?days="+tt.days)
			r.ServeHTTP(w, req)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func Test_GetDashboardHandler_shouldReturn400AndFieldSpecificMessage_whenLocalDateHeaderMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockGetDashboardUsecase(t)
	r := initDashboardRouter(ctx, t, usecase, fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, dashboardURL, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "X-Local-Date")
}

func Test_GetDashboardHandler_shouldReturn500_whenUsecaseFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	usecase := NewMockGetDashboardUsecase(t)
	usecase.On("GetDashboard", mock.Anything, mock.Anything).Return(nil, errors.New("firestore unavailable"))

	r := initDashboardRouter(ctx, t, usecase, fakeAuthMiddleware(fixtureUserID, "org1"), fakeOrgResolverMiddleware(fixtureOrganizationID))
	w := httptest.NewRecorder()

	// when
	req := newDashboardRequest(ctx, t, dashboardURL)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
