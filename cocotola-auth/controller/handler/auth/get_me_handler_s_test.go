package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

func Test_GetMeHandler_GetMe_shouldReturn200_whenAuthenticated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	settingFinder := newMockuserSettingFinder(t)
	setting, err := domain.NewUserSetting(fixtureAppUserID, 5, "ja")
	require.NoError(t, err)
	settingFinder.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(setting, nil)
	r := initAuthRouterWithDeps(ctx, t, authUsecase, fakeAuthMiddleware(fixtureAppUserID, "user42", "org1"), settingFinder)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/auth/me", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	userIDExpr := parseExpr(t, "$.userId")
	userID := userIDExpr.Get(jsonObj)
	require.Len(t, userID, 1)
	assert.Equal(t, fixtureAppUserID.String(), userID[0])

	loginIDExpr := parseExpr(t, "$.loginId")
	loginID := loginIDExpr.Get(jsonObj)
	require.Len(t, loginID, 1)
	assert.Equal(t, "user42", loginID[0])

	orgNameExpr := parseExpr(t, "$.organizationName")
	orgName := orgNameExpr.Get(jsonObj)
	require.Len(t, orgName, 1)
	assert.Equal(t, "org1", orgName[0])

	languageExpr := parseExpr(t, "$.language")
	language := languageExpr.Get(jsonObj)
	require.Len(t, language, 1)
	assert.Equal(t, "ja", language[0])
}

func Test_GetMeHandler_GetMe_shouldReturnDefaultLanguage_whenSettingNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	settingFinder := newMockuserSettingFinder(t)
	settingFinder.On("FindByAppUserID", mock.Anything, fixtureAppUserID).Return(nil, domain.ErrUserSettingNotFound)
	r := initAuthRouterWithDeps(ctx, t, authUsecase, fakeAuthMiddleware(fixtureAppUserID, "user42", "org1"), settingFinder)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/auth/me", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	languageExpr := parseExpr(t, "$.language")
	language := languageExpr.Get(jsonObj)
	require.Len(t, language, 1)
	assert.Equal(t, "en", language[0])
}

func Test_GetMeHandler_GetMe_shouldReturn401_whenUserIDMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authUsecase := NewMockAuthUsecase(t)
	r := initAuthRouterWithMiddleware(ctx, t, authUsecase, noopMiddleware())
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/auth/me", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	validateErrorResponse(t, respBytes, "unauthorized", "Unauthorized")
}
