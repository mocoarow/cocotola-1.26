package authz_test

import (
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
	domainrbac "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/rbac"

	authzhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/authz"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
)

func initPolicyRouter(_ context.Context, t *testing.T, policyAdder *MockUserPolicyAdder) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(context.Background(), serverConfig, domain.AppName)
	require.NoError(t, err)
	api := router.Group("api")
	v1 := api.Group("v1")
	internalAuth := v1.Group("internal/auth")

	handler := authzhandler.NewAddPolicyHandler(policyAdder)
	authzhandler.InitAuthzPolicyRouter(handler, internalAuth)

	return router
}

func Test_AddPolicyHandler_shouldReturn204_whenPolicyAdded(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domainrbac.ResourceWorkbook("wb-1")
	require.NoError(t, err)
	policyAdder := NewMockUserPolicyAdder(t)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrgID, fixtureAppUserID,
		domainrbac.ActionCreateQuestion(), wbResource, domainrbac.EffectAllow(),
	).Return(nil)
	r := initPolicyRouter(ctx, t, policyAdder)
	w := httptest.NewRecorder()

	// when
	body := `{"org":"` + fixtureOrgID.String() + `","user":"` + fixtureAppUserID.String() + `","action":"create_question","resource":"workbook:wb-1","effect":"allow"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/authz/policy", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func Test_AddPolicyHandler_shouldReturn400_whenBodyMalformed(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	policyAdder := NewMockUserPolicyAdder(t)
	r := initPolicyRouter(ctx, t, policyAdder)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/authz/policy", strings.NewReader("{invalid"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "invalid request body")
}

func Test_AddPolicyHandler_shouldReturn400_whenRequiredFieldsMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	policyAdder := NewMockUserPolicyAdder(t)
	r := initPolicyRouter(ctx, t, policyAdder)
	w := httptest.NewRecorder()

	// when
	body := `{"org":"` + fixtureOrgID.String() + `","user":"` + fixtureAppUserID.String() + `"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/authz/policy", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "org, user, action, resource, and effect are required")
}

func Test_AddPolicyHandler_shouldReturn400_whenOrgNotUUID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	policyAdder := NewMockUserPolicyAdder(t)
	r := initPolicyRouter(ctx, t, policyAdder)
	w := httptest.NewRecorder()

	// when
	body := `{"org":"not-uuid","user":"` + fixtureAppUserID.String() + `","action":"create_question","resource":"workbook:wb-1","effect":"allow"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/authz/policy", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "org must be a UUID")
}

func Test_AddPolicyHandler_shouldReturn400_whenEffectInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	policyAdder := NewMockUserPolicyAdder(t)
	r := initPolicyRouter(ctx, t, policyAdder)
	w := httptest.NewRecorder()

	// when
	body := `{"org":"` + fixtureOrgID.String() + `","user":"` + fixtureAppUserID.String() + `","action":"create_question","resource":"workbook:wb-1","effect":"maybe"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/authz/policy", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "effect must be 'allow' or 'deny'")
}

func Test_AddPolicyHandler_shouldReturn500_whenPolicyAdderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	wbResource, err := domainrbac.ResourceWorkbook("wb-1")
	require.NoError(t, err)
	policyAdder := NewMockUserPolicyAdder(t)
	policyAdder.On("AddPolicyForUser", mock.Anything, fixtureOrgID, fixtureAppUserID,
		domainrbac.ActionCreateQuestion(), wbResource, domainrbac.EffectAllow(),
	).Return(errors.New("db error"))
	r := initPolicyRouter(ctx, t, policyAdder)
	w := httptest.NewRecorder()

	// when
	body := `{"org":"` + fixtureOrgID.String() + `","user":"` + fixtureAppUserID.String() + `","action":"create_question","resource":"workbook:wb-1","effect":"allow"}`
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/internal/auth/authz/policy", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
