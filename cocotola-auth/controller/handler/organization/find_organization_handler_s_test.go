package organization_test

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

func Test_FindOrganizationHandler_FindOrganization_shouldReturn200_whenOrganizationExists(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	org := domain.ReconstructOrganization(1, "test-org", 100, 50)
	orgFinder := NewMockFinder(t)
	orgFinder.On("FindByName", mock.Anything, "test-org").Return(org, nil)
	r := initOrgRouter(ctx, t, orgFinder, fakeAuthMiddleware(42, "user42", "test-org"))
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/auth/organization?name=test-org", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusOK, w.Code)

	jsonObj := parseJSON(t, respBytes)
	idExpr := parseExpr(t, "$.id")
	id := idExpr.Get(jsonObj)
	require.Len(t, id, 1)
	assert.EqualValues(t, 1, id[0])

	nameExpr := parseExpr(t, "$.name")
	name := nameExpr.Get(jsonObj)
	require.Len(t, name, 1)
	assert.Equal(t, "test-org", name[0])
}

func Test_FindOrganizationHandler_FindOrganization_shouldReturn400_whenNameMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgFinder := NewMockFinder(t)
	r := initOrgRouter(ctx, t, orgFinder, fakeAuthMiddleware(42, "user42", "test-org"))
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/auth/organization", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
	validateErrorResponse(t, respBytes, "bad_request", "name query parameter is required")
}

func Test_FindOrganizationHandler_FindOrganization_shouldReturn404_whenOrganizationNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	orgFinder := NewMockFinder(t)
	orgFinder.On("FindByName", mock.Anything, "nonexistent").Return(nil, domain.ErrOrganizationNotFound)
	r := initOrgRouter(ctx, t, orgFinder, fakeAuthMiddleware(42, "user42", "test-org"))
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/auth/organization?name=nonexistent", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	assert.Equal(t, http.StatusNotFound, w.Code)
	validateErrorResponse(t, respBytes, "organization_not_found", "organization not found")
}
