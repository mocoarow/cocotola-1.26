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

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/middleware"
)

func setupAPIKeyRouter(t *testing.T, expectedKey string) *gin.Engine {
	t.Helper()

	r := gin.New()
	r.Use(middleware.NewAPIKeyMiddleware(expectedKey))
	r.GET("/internal/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func Test_APIKeyMiddleware_shouldReturn200_whenAPIKeyIsValid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := setupAPIKeyRouter(t, "test-secret-key-that-is-long-enough")
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Service-Api-Key", "test-secret-key-that-is-long-enough")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusOK, w.Code)
}

func Test_APIKeyMiddleware_shouldReturn401_whenAPIKeyIsMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := setupAPIKeyRouter(t, "test-secret-key-that-is-long-enough")
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_APIKeyMiddleware_shouldReturn403_whenAPIKeyIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	r := setupAPIKeyRouter(t, "test-secret-key-that-is-long-enough")
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	req.Header.Set("X-Service-Api-Key", "wrong-key-value-that-is-different")
	r.ServeHTTP(w, req)

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
}
