package usersetting_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

const updateTimezoneURL = "/api/v1/auth/user-setting/timezone"

func newTimezoneRequest(ctx context.Context, t *testing.T, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, updateTimezoneURL, bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func Test_UpdateTimezoneHandler_shouldReturn204_whenSettingExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	existing, err := domain.NewUserSetting(fixtureAppUserID, 5, "en", 10, "Asia/Tokyo")
	require.NoError(t, err)
	existing.SetVersion(2)
	saver := newMockuserSettingFinderSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.UserSetting) bool {
		return s.AppUserID() == fixtureAppUserID && s.Timezone() == "America/Los_Angeles" && s.Version() == 2
	})).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newTimezoneRequest(ctx, t, `{"timezone":"America/Los_Angeles"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UpdateTimezoneHandler_shouldCreateDefault_whenSettingNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingFinderSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, domain.ErrUserSettingNotFound)
	saver.On("Save", mock.Anything, mock.MatchedBy(func(s *domain.UserSetting) bool {
		return s.AppUserID() == fixtureAppUserID && s.Timezone() == "UTC" && s.Version() == 0
	})).Return(nil)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newTimezoneRequest(ctx, t, `{"timezone":"UTC"}`)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_UpdateTimezoneHandler_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	saver := newMockuserSettingFinderSaver(t)
	noopMiddleware := func(c *gin.Context) { c.Next() }
	r := initExternalUserSettingRouter(ctx, t, saver, noopMiddleware)
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
			saver := newMockuserSettingFinderSaver(t)
			r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
			w := httptest.NewRecorder()

			// when
			req := newTimezoneRequest(ctx, t, tt.body)
			r.ServeHTTP(w, req)
			respBytes := readBytes(t, w.Body)

			// then
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, string(respBytes), "invalid_request")
		})
	}
}

func Test_UpdateTimezoneHandler_shouldReturn400_whenTimezoneIsNotResolvable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	existing, err := domain.NewUserSetting(fixtureAppUserID, 5, "en", 10, "Asia/Tokyo")
	require.NoError(t, err)
	saver := newMockuserSettingFinderSaver(t)
	saver.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(existing, nil)
	r := initExternalUserSettingRouter(ctx, t, saver, fakeAuthMiddleware(fixtureAppUserID))
	w := httptest.NewRecorder()

	// when
	req := newTimezoneRequest(ctx, t, `{"timezone":"Not/AZone"}`)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, string(respBytes), "invalid_request")
}
