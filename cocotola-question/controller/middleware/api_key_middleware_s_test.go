//go:build small

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/middleware"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	testAPIKey   = "test-secret-key-that-is-long-enough"
	testWrongKey = "wrong-key-value-that-is-different"
	testOrgID    = "org-1"
)

type capturedContext struct {
	UserID         string
	OrganizationID string
}

func setupAPIKeyRouter(t *testing.T, expectedKey string, captured *capturedContext) *gin.Engine {
	t.Helper()

	r := gin.New()
	r.Use(middleware.NewAPIKeyMiddleware(expectedKey))
	r.GET("/internal/test", func(c *gin.Context) {
		if captured != nil {
			captured.UserID = c.GetString(controller.ContextFieldUserID{})
			captured.OrganizationID = c.GetString(controller.ContextFieldOrganizationID{})
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func Test_APIKeyMiddleware_shouldReturn200_whenAPIKeyAndOrgIDAreValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	captured := &capturedContext{}
	r := setupAPIKeyRouter(t, testAPIKey, captured)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Service-Api-Key", testAPIKey)
	req.Header.Set("X-Organization-Id", testOrgID)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, domain.SystemAppUserID, captured.UserID)
	assert.Equal(t, testOrgID, captured.OrganizationID)
}

func Test_APIKeyMiddleware_shouldReturn401_whenAPIKeyIsMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := setupAPIKeyRouter(t, testAPIKey, nil)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Organization-Id", testOrgID)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_APIKeyMiddleware_shouldReturn403_whenAPIKeyIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := setupAPIKeyRouter(t, testAPIKey, nil)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Service-Api-Key", testWrongKey)
	req.Header.Set("X-Organization-Id", testOrgID)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_APIKeyMiddleware_shouldReturn400_whenOrganizationIDIsMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := setupAPIKeyRouter(t, testAPIKey, nil)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Service-Api-Key", testAPIKey)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
