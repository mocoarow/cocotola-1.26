package authz_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"
)

func Test_CheckHandler_Check_shouldReturnAllowedTrue_whenUserHasPermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authzChecker := NewMockAuthorizationChecker(t)
	authzChecker.On("IsAllowed", mock.Anything, fixtureOrgID, fixtureAppUserID, domainrbac.ActionCreateWorkbook(), domainrbac.ResourceAny()).Return(true, nil)
	r := initAuthzRouter(ctx, t, authzChecker, fakeAuthMiddleware(fixtureAppUserID, "user42", "test-org"))
	w := httptest.NewRecorder()

	// when
	url := "/api/v1/auth/authz/check?org=" + fixtureOrgID.String() + "&user=" + fixtureAppUserID.String() + "&action=create_workbook&resource=*"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	allowedExpr := parseExpr(t, "$.allowed")
	allowed := allowedExpr.Get(jsonObj)
	require.Len(t, allowed, 1)
	assert.Equal(t, true, allowed[0])
}

func Test_CheckHandler_Check_shouldReturnAllowedFalse_whenUserLacksPermission(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authzChecker := NewMockAuthorizationChecker(t)
	authzChecker.On("IsAllowed", mock.Anything, fixtureOrgID, fixtureOtherID, domainrbac.ActionCreateWorkbook(), domainrbac.ResourceAny()).Return(false, nil)
	r := initAuthzRouter(ctx, t, authzChecker, fakeAuthMiddleware(fixtureOtherID, "user99", "test-org"))
	w := httptest.NewRecorder()

	// when
	url := "/api/v1/auth/authz/check?org=" + fixtureOrgID.String() + "&user=" + fixtureOtherID.String() + "&action=create_workbook&resource=*"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	allowedExpr := parseExpr(t, "$.allowed")
	allowed := allowedExpr.Get(jsonObj)
	require.Len(t, allowed, 1)
	assert.Equal(t, false, allowed[0])
}

func Test_CheckHandler_Check_shouldReturn400_whenOrgMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authzChecker := NewMockAuthorizationChecker(t)
	r := initAuthzRouter(ctx, t, authzChecker, fakeAuthMiddleware(fixtureAppUserID, "user42", "test-org"))
	w := httptest.NewRecorder()

	// when
	url := "/api/v1/auth/authz/check?user=" + fixtureAppUserID.String() + "&action=create_workbook&resource=*"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "org, user, action, and resource query parameters are required")
}

func Test_CheckHandler_Check_shouldReturn400_whenOrgNotUUID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	authzChecker := NewMockAuthorizationChecker(t)
	r := initAuthzRouter(ctx, t, authzChecker, fakeAuthMiddleware(fixtureAppUserID, "user42", "test-org"))
	w := httptest.NewRecorder()

	// when
	url := "/api/v1/auth/authz/check?org=abc&user=" + fixtureAppUserID.String() + "&action=create_workbook&resource=*"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "org must be a UUID")
}
