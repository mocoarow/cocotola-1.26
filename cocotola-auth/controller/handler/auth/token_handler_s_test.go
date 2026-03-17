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

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	authservice "github.com/mocoarow/cocotola-1.26/cocotola-auth/service/auth"
)

func Test_RefreshHandler_Refresh_shouldReturn200_whenValidRefreshToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	refreshOutput, err := authservice.NewRefreshAccessTokenOutput("new-jwt-access-token")
	require.NoError(t, err)
	authUsecase.On("RefreshAccessToken", mock.Anything, mock.Anything).Return(refreshOutput, nil).Once()
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"refreshToken":"valid-refresh-token"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
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
	assert.Equal(t, "new-jwt-access-token", accessToken[0])
}

func Test_RefreshHandler_Refresh_shouldReturn401_whenRefreshTokenExpired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("RefreshAccessToken", mock.Anything, mock.Anything).Return(nil, domain.ErrSessionExpired).Once()
	r := initAuthRouter(t, ctx, authUsecase)
	w := httptest.NewRecorder()
	body := `{"refreshToken":"expired-token"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/auth/refresh", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "invalid_refresh_token", "refresh token is invalid or expired")
}
