//go:build small

package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/middleware"
)

const testOrganizationName = "org-name-42"

type organizationResolverResult struct {
	recorder       *httptest.ResponseRecorder
	organizationID string
	reachedHandler bool
}

func serveWithOrganizationResolverMiddleware(t *testing.T, resolver middleware.OrganizationIDResolver, organizationName string) organizationResolverResult {
	t.Helper()

	var result organizationResolverResult
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if organizationName != "" {
			c.Set(controller.ContextFieldOrganizationName{}, organizationName)
		}
		c.Next()
	})
	r.Use(middleware.NewOrganizationResolverMiddleware(resolver))
	r.GET("/protected", func(c *gin.Context) {
		result.reachedHandler = true
		result.organizationID = c.GetString(controller.ContextFieldOrganizationID{})
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/protected", nil)
	require.NoError(t, err)
	result.recorder = httptest.NewRecorder()
	r.ServeHTTP(result.recorder, req)
	return result
}

func Test_NewOrganizationResolverMiddleware_shouldReturn401_whenOrganizationNameIsMissing(t *testing.T) {
	t.Parallel()

	// given
	resolver := NewMockOrganizationIDResolver(t)

	// when
	result := serveWithOrganizationResolverMiddleware(t, resolver, "")

	// then
	assert.Equal(t, http.StatusUnauthorized, result.recorder.Code)
}

func Test_NewOrganizationResolverMiddleware_shouldNotReachHandler_whenOrganizationNameIsMissing(t *testing.T) {
	t.Parallel()

	// given
	resolver := NewMockOrganizationIDResolver(t)

	// when
	result := serveWithOrganizationResolverMiddleware(t, resolver, "")

	// then
	assert.False(t, result.reachedHandler)
}

func Test_NewOrganizationResolverMiddleware_shouldReturn500_whenResolveFails(t *testing.T) {
	t.Parallel()

	// given
	resolver := NewMockOrganizationIDResolver(t)
	resolver.On("ResolveOrganizationID", mock.Anything, testOrganizationName).Return("", errors.New("auth service down"))

	// when
	result := serveWithOrganizationResolverMiddleware(t, resolver, testOrganizationName)

	// then
	assert.Equal(t, http.StatusInternalServerError, result.recorder.Code)
}

func Test_NewOrganizationResolverMiddleware_shouldNotReachHandler_whenResolveFails(t *testing.T) {
	t.Parallel()

	// given
	resolver := NewMockOrganizationIDResolver(t)
	resolver.On("ResolveOrganizationID", mock.Anything, testOrganizationName).Return("", errors.New("auth service down"))

	// when
	result := serveWithOrganizationResolverMiddleware(t, resolver, testOrganizationName)

	// then
	assert.False(t, result.reachedHandler)
}

func Test_NewOrganizationResolverMiddleware_shouldSetOrganizationIDInContext_whenResolveSucceeds(t *testing.T) {
	t.Parallel()

	// given
	resolver := NewMockOrganizationIDResolver(t)
	resolver.On("ResolveOrganizationID", mock.Anything, testOrganizationName).Return("org-id-42", nil)

	// when
	result := serveWithOrganizationResolverMiddleware(t, resolver, testOrganizationName)

	// then
	assert.Equal(t, "org-id-42", result.organizationID)
}

func Test_OrganizationIDResolverFunc_ResolveOrganizationID_shouldReturnFunctionResult_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	resolve := middleware.OrganizationIDResolverFunc(func(_ context.Context, name string) (string, error) {
		return "id-of-" + name, nil
	})

	// when
	id, err := resolve.ResolveOrganizationID(context.Background(), testOrganizationName)

	// then
	require.NoError(t, err)
	assert.Equal(t, "id-of-"+testOrganizationName, id)
}
