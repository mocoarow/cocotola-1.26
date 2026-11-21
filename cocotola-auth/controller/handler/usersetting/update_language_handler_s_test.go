package usersetting_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	usersettingservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/usersetting"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

const updateLanguageURL = "/api/v1/auth/user-setting/language"

func newAuthedRequest(ctx context.Context, t *testing.T, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, updateLanguageURL, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func Test_UpdateLanguageHandler_shouldPassUserIDAndLanguageToCommand_whenRequestIsValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateLanguage", mock.Anything, &usersettingservice.UpdateLanguageInput{AppUserID: fixtureAppUserID, Language: "ja"}).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UpdateLanguageHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	noopMiddleware := func(c *gin.Context) { c.Next() }
	r := initExternalUserSettingRouter(ctx, t, updater, noopMiddleware)
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", http.StatusText(http.StatusUnauthorized))
}

func Test_UpdateLanguageHandler_shouldReturn400_whenBodyInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tests := []struct {
		name string
		body string
	}{
		{name: "notJSON", body: `not-json`},
		{name: "missingField", body: `{}`},
		{name: "wrongLength", body: `{"language":"jpn"}`},
		{name: "outsideWhitelist", body: `{"language":"fr"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			updater := newMockuserSettingUpdater(t)
			r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
			w := httptest.NewRecorder()

			// when
			req := newAuthedRequest(ctx, t, tt.body)
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
		})
	}
}

func Test_UpdateLanguageHandler_shouldReturn400_whenCommandRejectsValue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateLanguage", mock.Anything, mock.Anything).Return(fmt.Errorf("change: %w", usersettingservice.ErrInvalidValue))
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "language is invalid")
}

func Test_UpdateLanguageHandler_shouldReturn409_whenConcurrentModification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateLanguage", mock.Anything, mock.Anything).Return(fmt.Errorf("save: %w", libversioned.ErrConcurrentModification))
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusConflict, w.Code)
	validateErrorResponse(t, respBytes, "conflict", "user setting was modified concurrently")
}

func Test_UpdateLanguageHandler_shouldReturn404_whenSettingDeletedConcurrently(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateLanguage", mock.Anything, mock.Anything).Return(fmt.Errorf("save: %w", domain.ErrUserSettingNotFound))
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "user_setting_not_found", "user setting not found")
}

func Test_UpdateLanguageHandler_shouldReturn500_whenCommandFailsUnexpectedly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	updater.On("UpdateLanguage", mock.Anything, mock.Anything).Return(errors.New("db connection lost"))
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	validateErrorResponse(t, respBytes, "internal_server_error", http.StatusText(http.StatusInternalServerError))
}

func Test_UpdateLanguageHandler_shouldRouteOnlyToPUT(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	updater := newMockuserSettingUpdater(t)
	r := initExternalUserSettingRouter(ctx, t, updater, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when (POST is not registered)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, updateLanguageURL, strings.NewReader(`{"language":"ja"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
}
