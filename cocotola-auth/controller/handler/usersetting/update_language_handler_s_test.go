package usersetting_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
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

func Test_UpdateLanguageHandler_shouldReturn204_whenSettingExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	existing, err := domain.NewUserSetting(fixtureAppUserID, 5, "en")
	require.NoError(t, err)
	existing.SetVersion(7) // simulate a row already persisted with version=7
	saver := newMockuserSettingSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.UserSetting) bool {
		return s.AppUserID() == fixtureAppUserID && s.Language() == "ja" && s.Version() == 7
	})).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UpdateLanguageHandler_shouldCreateDefault_whenSettingNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, domain.ErrUserSettingNotFound)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.UserSetting) bool {
		// version==0 confirms the handler took the INSERT path (default-init).
		return s.AppUserID() == fixtureAppUserID && s.Language() == "ja" && s.Version() == 0
	})).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
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
	saver := newMockuserSettingSaver(t)
	noopMiddleware := func(c *gin.Context) { c.Next() }
	r := initExternalUserSettingRouter(ctx, t, saver, noopMiddleware)
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

	// given
	saver := newMockuserSettingSaver(t)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `not-json`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
}

func Test_UpdateLanguageHandler_shouldReturn400_whenLanguageMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingSaver(t)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
}

func Test_UpdateLanguageHandler_shouldReturn400_whenLanguageWrongLength(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingSaver(t)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when (3-letter code rejected by binding:"len=2")
	req := newAuthedRequest(ctx, t, `{"language":"jpn"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func Test_UpdateLanguageHandler_shouldReturn400_whenLanguageOutsideWhitelist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given (binding accepts only en/ja/ko; valid ISO 639-1 codes outside the
	// whitelist must be rejected at the API boundary before reaching the repo)
	saver := newMockuserSettingSaver(t)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"fr"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_request", "request body is invalid")
}

func Test_UpdateLanguageHandler_shouldReturn409_whenConcurrentModification(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	existing, err := domain.NewUserSetting(fixtureAppUserID, 5, "en")
	require.NoError(t, err)
	saver := newMockuserSettingSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	saver.On("Save", mock.Anything, mock.Anything).Return(libversioned.ErrConcurrentModification)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
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

	// given: caller loaded a setting at version 7, but the row was deleted before save
	existing, err := domain.NewUserSetting(fixtureAppUserID, 5, "en")
	require.NoError(t, err)
	existing.SetVersion(7)
	saver := newMockuserSettingSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	saver.On("Save", mock.Anything, mock.Anything).Return(domain.ErrUserSettingNotFound)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "user_setting_not_found", "user setting not found")
}

func Test_UpdateLanguageHandler_shouldReturn500_whenFindFailsUnexpectedly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, errors.New("db connection lost"))
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_UpdateLanguageHandler_shouldReturn500_whenSaveFailsUnexpectedly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	existing, err := domain.NewUserSetting(fixtureAppUserID, 5, "en")
	require.NoError(t, err)
	saver := newMockuserSettingSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	saver.On("Save", mock.Anything, mock.Anything).Return(errors.New("db connection lost"))
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newAuthedRequest(ctx, t, `{"language":"ja"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_UpdateLanguageHandler_shouldRouteOnlyToPUT(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingSaver(t)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when (POST is not registered)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, updateLanguageURL, strings.NewReader(`{"language":"ja"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
}
