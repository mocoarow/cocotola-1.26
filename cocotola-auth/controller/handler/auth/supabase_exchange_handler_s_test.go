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

func Test_SupabaseExchangeHandler_Exchange_shouldReturn200_whenTokenIsValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	exchangeOutput, err := authservice.NewSupabaseExchangeOutput(1, "user@example.com", "org1")
	require.NoError(t, err)
	authUsecase.On("SupabaseExchange", mock.Anything, mock.Anything).Return(exchangeOutput, nil).Once()

	tokenOutput, err := authservice.NewCreateTokenPairOutput("jwt-access-token", "opaque-refresh-token")
	require.NoError(t, err)
	authUsecase.On("CreateTokenPair", mock.Anything, mock.Anything).Return(tokenOutput, nil).Once()

	r := initAuthRouter(ctx, t, authUsecase)
	w := httptest.NewRecorder()
	body := `{"supabaseJwt":"valid-jwt","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/supabase/exchange", strings.NewReader(body))
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

func Test_SupabaseExchangeHandler_Exchange_shouldReturn400_whenRequestBodyIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	r := initAuthRouter(ctx, t, authUsecase)
	w := httptest.NewRecorder()
	body := `{"supabaseJwt":""}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/supabase/exchange", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "invalid_supabase_exchange_request", "request body is invalid")
}

func Test_SupabaseExchangeHandler_Exchange_shouldReturn401_whenTokenIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("SupabaseExchange", mock.Anything, mock.Anything).Return(nil, errors.New("verify supabase token: invalid")).Once()

	r := initAuthRouter(ctx, t, authUsecase)
	w := httptest.NewRecorder()
	body := `{"supabaseJwt":"invalid-jwt","organizationName":"org1"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/supabase/exchange", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthenticated", "invalid supabase token")
}

func Test_SupabaseExchangeHandler_Exchange_shouldReturn400_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	authUsecase.On("SupabaseExchange", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("find organization unknown-org: %w", domain.ErrOrganizationNotFound)).Once()

	r := initAuthRouter(ctx, t, authUsecase)
	w := httptest.NewRecorder()
	body := `{"supabaseJwt":"valid-jwt","organizationName":"unknown-org"}`

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/supabase/exchange", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "organization_not_found", "organization not found")
}
