package usersetting_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"
)

const updateTimezoneURL = "/api/v1/auth/user-setting/timezone"

func newTimezoneRequest(ctx context.Context, t *testing.T, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, updateTimezoneURL, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func Test_UpdateTimezoneHandler_shouldPassUserIDAndTimezoneToCommand_whenRequestIsValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateTimezone", mock.Anything, &usersettingservice.UpdateTimezoneInput{AppUserID: fixtureAppUserID, Timezone: "America/Los_Angeles"}).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newTimezoneRequest(ctx, t, `{"timezone":"America/Los_Angeles"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UpdateTimezoneHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	noopMiddleware := func(c *gin.Context) { c.Next() }
	r := initExternalUserSettingRouter(ctx, t, updater, noopMiddleware)
	w := httptest.NewRecorder()

	// when
	req := newTimezoneRequest(ctx, t, `{"timezone":"UTC"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", http.StatusText(http.StatusUnauthorized))
}

func Test_UpdateTimezoneHandler_shouldReturn400_whenBodyInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		body string
	}{
		{name: "notJSON", body: "not json"},
		{name: "missingField", body: `{}`},
		{name: "emptyString", body: `{"timezone":""}`},
		{name: "tooLong", body: `{"timezone":"` + strings.Repeat("a", 65) + `"}`},
		{name: "disallowedSpace", body: `{"timezone":"Asia Tokyo"}`},
		{name: "disallowedQuestionMark", body: `{"timezone":"Asia/Tokyo?"}`},
		{name: "disallowedNonAscii", body: `{"timezone":"アジア/東京"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			updater := newMockuserSettingUpdater(t)
			r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
			w := httptest.NewRecorder()

			// when
			req := newTimezoneRequest(ctx, t, tt.body)
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
		})
	}
}

func Test_UpdateTimezoneHandler_shouldReturn400_whenCommandRejectsValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateTimezone", mock.Anything, mock.Anything).Return(fmt.Errorf("change: %w", usersettingservice.ErrInvalidValue))
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newTimezoneRequest(ctx, t, `{"timezone":"Not/AZone"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "timezone is invalid")
}
