package auth_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

func Test_PasswordAuthenticateHandler_Authenticate_shouldReturn200_whenValidCredentialsWithJSON(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authOutput, err := authservice.NewPasswordAuthenticateOutput(1, "user1", "org1")
	require.NoError(t, err)
	authUsecase.On("PasswordAuthenticate", mock.Anything, mock.Anything).Return(authOutput, nil).Once()

	tokenOutput, err := authservice.NewCreateTokenPairOutput("jwt-access-token", "opaque-refresh-token")
	require.NoError(t, err)
	authUsecase.On("CreateTokenPair", mock.Anything, mock.Anything).Return(tokenOutput, nil).Once()

	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"loginId":"user1","password":"password1","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/authenticate", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	accessTokenExpr := parseExpr(t, "$.accessToken")
	accessToken := accessTokenExpr.Get(jsonObj)
	require.Len(t, accessToken, 1)
	assert.Equal(t, "jwt-access-token", accessToken[0])

	refreshTokenExpr := parseExpr(t, "$.refreshToken")
	refreshToken := refreshTokenExpr.Get(jsonObj)
	require.Len(t, refreshToken, 1)
	assert.Equal(t, "opaque-refresh-token", refreshToken[0])
}

func Test_PasswordAuthenticateHandler_Authenticate_shouldReturn400_whenRequestBodyIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"loginId":"","password":""}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/authenticate", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_authenticate_request", "request body is invalid")
}

func Test_PasswordAuthenticateHandler_Authenticate_shouldReturn401_whenUsecaseReturnsErrUnauthenticated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("PasswordAuthenticate", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("authenticate user: %w", domain.ErrUnauthenticated)).Once()
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"loginId":"user1","password":"password1","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/authenticate", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthenticated", "Unauthorized")
}

func Test_PasswordAuthenticateHandler_Authenticate_shouldReturn500_whenUsecaseReturnsUnexpectedError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("PasswordAuthenticate", mock.Anything, mock.Anything).Return(nil, errors.New("unexpected error")).Once()
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"loginId":"user1","password":"password1","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/authenticate", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	validateErrorResponse(t, respBytes, "internal_server_error", "Internal Server Error")
}

func Test_PasswordAuthenticateHandler_Authenticate_shouldSetCookie_whenXTokenDeliveryIsCookie(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authOutput, err := authservice.NewPasswordAuthenticateOutput(1, "user1", "org1")
	require.NoError(t, err)
	authUsecase.On("PasswordAuthenticate", mock.Anything, mock.Anything).Return(authOutput, nil).Once()

	sessionOutput, err := authservice.NewCreateSessionTokenOutput("raw-session-token")
	require.NoError(t, err)
	authUsecase.On("CreateSessionToken", mock.Anything, mock.Anything).Return(sessionOutput, nil).Once()

	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"loginId":"user1","password":"password1","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/authenticate", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token-Delivery", "cookie")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	sessionCookie := findCookieByName(t, w.Result().Cookies(), "session_token")
	require.NotNil(t, sessionCookie, "session_token cookie should be set")
	assert.Equal(t, "raw-session-token", sessionCookie.Value)
	assert.True(t, sessionCookie.HttpOnly)
	assert.Equal(t, "/", sessionCookie.Path)
	assert.Equal(t, 1800, sessionCookie.MaxAge)
}

func Test_PasswordAuthenticateHandler_Authenticate_shouldReturn400_whenXTokenDeliveryIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"loginId":"user1","password":"password1","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/authenticate", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Token-Delivery", "invalid-value")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_token_delivery", "X-Token-Delivery must be 'json' or 'cookie'")
}
