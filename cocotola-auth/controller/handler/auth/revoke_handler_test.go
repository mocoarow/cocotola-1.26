package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_RevokeHandler_Logout_shouldReturn204_andClearCookie(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("RevokeSessionToken", mock.Anything, mock.Anything).Return(nil).Maybe()
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/logout", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)

	sessionCookie := findCookieByName(t, w.Result().Cookies(), "session_token")
	require.NotNil(t, sessionCookie, "session_token cookie should be set")
	assert.Empty(t, sessionCookie.Value)
	assert.Equal(t, -1, sessionCookie.MaxAge)
	assert.True(t, sessionCookie.HttpOnly)
	assert.Equal(t, "/", sessionCookie.Path)
}

func Test_RevokeHandler_Revoke_shouldReturn204_whenTokenRevoked(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("RevokeToken", mock.Anything, mock.Anything).Return(nil).Once()
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"token":"some-token"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/revoke", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}
