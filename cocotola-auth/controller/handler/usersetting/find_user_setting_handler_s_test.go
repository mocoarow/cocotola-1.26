package usersetting_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_FindUserSettingHandler_shouldReturn200_whenSettingExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	setting, err := domain.NewUserSetting(fixtureAppUserID, 5, "ja", 25, "America/Los_Angeles")
	require.NoError(t, err)
	settingFinder := newMockuserSettingFinder(t)
	settingFinder.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(setting, nil)
	r := initUserSettingRouter(ctx, t, settingFinder)
	w := httptest.NewRecorder()

	// when
	reqURL := "/api/v1/internal/auth/user-setting?user_id=" + fixtureAppUserID.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)

	maxWbExpr := parseExpr(t, "$.maxWorkbooks")
	maxWb := maxWbExpr.Get(jsonObj)
	require.Len(t, maxWb, 1)
	assert.EqualValues(t, 5, maxWb[0])

	// Persisted UserSetting overrides defaults — verify the non-default
	// values flow through to the response so a regression in the field
	// mapping cannot hide behind a default-equal value.
	languageExpr := parseExpr(t, "$.language")
	language := languageExpr.Get(jsonObj)
	require.Len(t, language, 1)
	assert.Equal(t, "ja", language[0])

	dailyGoalExpr := parseExpr(t, "$.dailyGoal")
	dailyGoal := dailyGoalExpr.Get(jsonObj)
	require.Len(t, dailyGoal, 1)
	assert.EqualValues(t, 25, dailyGoal[0])

	timezoneExpr := parseExpr(t, "$.timezone")
	timezone := timezoneExpr.Get(jsonObj)
	require.Len(t, timezone, 1)
	assert.Equal(t, "America/Los_Angeles", timezone[0])
}

func Test_FindUserSettingHandler_shouldReturnDefaults_whenSettingNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	settingFinder := newMockuserSettingFinder(t)
	settingFinder.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, domain.ErrUserSettingNotFound)
	r := initUserSettingRouter(ctx, t, settingFinder)
	w := httptest.NewRecorder()

	// when
	reqURL := "/api/v1/internal/auth/user-setting?user_id=" + fixtureAppUserID.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	// Pin every defaulted field here so the "internal user-setting GET
	// returns sensible defaults when no row exists" contract has a single
	// canonical assertion. Any drift between domain defaults and the
	// values surfaced to the internal caller breaks this test.
	jsonObj := parseJSON(t, respBytes)

	maxWbExpr := parseExpr(t, "$.maxWorkbooks")
	maxWb := maxWbExpr.Get(jsonObj)
	require.Len(t, maxWb, 1)
	assert.EqualValues(t, 3, maxWb[0])

	languageExpr := parseExpr(t, "$.language")
	language := languageExpr.Get(jsonObj)
	require.Len(t, language, 1)
	assert.Equal(t, domain.DefaultLanguage(), language[0])

	dailyGoalExpr := parseExpr(t, "$.dailyGoal")
	dailyGoal := dailyGoalExpr.Get(jsonObj)
	require.Len(t, dailyGoal, 1)
	assert.EqualValues(t, domain.DefaultDailyGoal(), dailyGoal[0])

	timezoneExpr := parseExpr(t, "$.timezone")
	timezone := timezoneExpr.Get(jsonObj)
	require.Len(t, timezone, 1)
	assert.Equal(t, domain.DefaultTimezone(), timezone[0])
}

func Test_FindUserSettingHandler_shouldReturn400_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	settingFinder := newMockuserSettingFinder(t)
	r := initUserSettingRouter(ctx, t, settingFinder)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/internal/auth/user-setting", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "user_id query parameter is required")
}

func Test_FindUserSettingHandler_shouldReturn400_whenUserIDInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	settingFinder := newMockuserSettingFinder(t)
	r := initUserSettingRouter(ctx, t, settingFinder)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/internal/auth/user-setting?user_id=invalid", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "invalid user_id")
}

func Test_FindUserSettingHandler_shouldReturn500_whenFinderReturnsUnexpectedError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	settingFinder := newMockuserSettingFinder(t)
	settingFinder.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, errors.New("db connection lost"))
	r := initUserSettingRouter(ctx, t, settingFinder)
	w := httptest.NewRecorder()

	// when
	reqURL := "/api/v1/internal/auth/user-setting?user_id=" + fixtureAppUserID.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
