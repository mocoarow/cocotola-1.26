package usersetting_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

const updateDailyGoalURL = "/api/v1/auth/user-setting/daily-goal"

func newDailyGoalRequest(ctx context.Context, t *testing.T, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, updateDailyGoalURL, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func Test_UpdateDailyGoalHandler_shouldPassUserIDAndDailyGoalToCommand_whenRequestIsValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateDailyGoal", mock.Anything, &usersettingservice.UpdateDailyGoalInput{AppUserID: fixtureAppUserID, DailyGoal: 25}).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newDailyGoalRequest(ctx, t, `{"dailyGoal":25}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UpdateDailyGoalHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	noopMiddleware := func(c *gin.Context) { c.Next() }
	r := initExternalUserSettingRouter(ctx, t, updater, noopMiddleware)
	w := httptest.NewRecorder()

	// when
	req := newDailyGoalRequest(ctx, t, `{"dailyGoal":25}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", http.StatusText(http.StatusUnauthorized))
}

func Test_UpdateDailyGoalHandler_shouldReturn400_whenBodyInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		body string
	}{
		{name: "notJSON", body: "not json"},
		{name: "missingField", body: `{}`},
		{name: "belowMin", body: `{"dailyGoal":0}`},
		{name: "aboveMax", body: `{"dailyGoal":501}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			updater := newMockuserSettingUpdater(t)
			r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
			w := httptest.NewRecorder()

			// when
			req := newDailyGoalRequest(ctx, t, tt.body)
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
		})
	}
}

func Test_UpdateDailyGoalHandler_shouldReturn400_whenCommandRejectsValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateDailyGoal", mock.Anything, mock.Anything).Return(fmt.Errorf("change: %w", usersettingservice.ErrInvalidValue))
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newDailyGoalRequest(ctx, t, `{"dailyGoal":25}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "daily goal is invalid")
}
