package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/middleware"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

var testCookieConfig = controller.CookieConfig{
	Name:     "session_token",
	Path:     "/",
	Secure:   false,
	SameSite: "Lax",
}

func setupRouter(t *testing.T, authUsecase middleware.AuthUsecase) *gin.Engine {
	t.Helper()
	r := gin.New()
	r.Use(middleware.NewAuthMiddleware(authUsecase, testCookieConfig, 30))
	r.GET("/protected", func(c *gin.Context) {
		userID := c.GetInt(controller.ContextFieldUserID{})
		c.JSON(http.StatusOK, gin.H{"userId": userID})
	})
	return r
}

func Test_AuthMiddleware_shouldReturn200AndSetUserID_whenValidBearerToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	output, err := authservice.NewValidateAccessTokenOutput(42, "user42", "org1")
	require.NoError(t, err)

	mockUsecase := NewMockAuthUsecase(t)
	mockUsecase.On("ValidateAccessToken", mock.Anything, mock.Anything).Return(output, nil).Once()
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer valid-token")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"userId":42`)
}

func Test_AuthMiddleware_shouldReturn401_whenNoTokenProvided(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	mockUsecase := NewMockAuthUsecase(t)
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_AuthMiddleware_shouldReturn401_whenAuthorizationHeaderIsNotBearer(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	mockUsecase := NewMockAuthUsecase(t)
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_AuthMiddleware_shouldReturn401_whenValidateAccessTokenReturnsError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	mockUsecase := NewMockAuthUsecase(t)
	mockUsecase.On("ValidateAccessToken", mock.Anything, mock.Anything).Return(nil, errors.New("invalid token")).Once()
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_AuthMiddleware_shouldReturn200_whenValidSessionCookie(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	output, err := authservice.NewValidateSessionTokenOutput(42, "user42", "org1")
	require.NoError(t, err)

	mockUsecase := NewMockAuthUsecase(t)
	mockUsecase.On("ValidateSessionToken", mock.Anything, mock.Anything).Return(output, nil).Once()
	mockUsecase.On("ExtendSessionToken", mock.Anything, mock.Anything).Return(nil).Once()
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "valid-session-token"})
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"userId":42`)
}

func Test_AuthMiddleware_shouldSetNewCookie_whenSessionTokenExtended(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	output, err := authservice.NewValidateSessionTokenOutput(42, "user42", "org1")
	require.NoError(t, err)

	mockUsecase := NewMockAuthUsecase(t)
	mockUsecase.On("ValidateSessionToken", mock.Anything, mock.Anything).Return(output, nil).Once()
	mockUsecase.On("ExtendSessionToken", mock.Anything, mock.Anything).Return(nil).Once()
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "session-token-value"})
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	// verify cookie is refreshed with sliding window
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}
	require.NotNil(t, sessionCookie, "session_token cookie should be set with extended expiry")
	assert.Equal(t, "session-token-value", sessionCookie.Value)
	assert.True(t, sessionCookie.HttpOnly)
	assert.Equal(t, 1800, sessionCookie.MaxAge)
}

func Test_AuthMiddleware_shouldPreferBearerOverCookie(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	output, err := authservice.NewValidateAccessTokenOutput(42, "user42", "org1")
	require.NoError(t, err)

	mockUsecase := NewMockAuthUsecase(t)
	mockUsecase.On("ValidateAccessToken", mock.Anything, mock.Anything).Return(output, nil).Once()
	// No ValidateSessionToken or ExtendSessionToken calls expected because Bearer takes priority
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer bearer-token")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "cookie-token"})
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	// verify no cookie refresh happened (no Set-Cookie header)
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		assert.NotEqual(t, "session_token", c.Name)
	}
}

func Test_AuthMiddleware_shouldReturn401_whenSessionCookieIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	mockUsecase := NewMockAuthUsecase(t)
	mockUsecase.On("ValidateSessionToken", mock.Anything, mock.Anything).Return(nil, errors.New("session expired")).Once()
	r := setupRouter(t, mockUsecase)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-session"})
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
