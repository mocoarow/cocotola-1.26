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

func setupRouter(t *testing.T, mw gin.HandlerFunc, captured *capturedContext) *gin.Engine {
	t.Helper()

	r := gin.New()
	r.Use(mw)
	r.GET("/internal/test", func(c *gin.Context) {
		if captured != nil {
			captured.UserID = c.GetString(controller.ContextFieldUserID{})
			captured.OrganizationID = c.GetString(controller.ContextFieldOrganizationID{})
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}

func serve(t *testing.T, r *gin.Engine, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/test", nil)
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func Test_APIKeyMiddleware_shouldSetSystemUserID_whenAPIKeyIsValidWithoutOrganizationID(t *testing.T) {
	t.Parallel()

	// given
	captured := &capturedContext{}
	r := setupRouter(t, middleware.NewAPIKeyMiddleware(testAPIKey), captured)

	// when
	w := serve(t, r, map[string]string{"X-Service-Api-Key": testAPIKey})

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, domain.SystemAppUserID, captured.UserID)
}

func Test_APIKeyMiddleware_shouldReturn401_whenAPIKeyIsMissing(t *testing.T) {
	t.Parallel()

	// given
	r := setupRouter(t, middleware.NewAPIKeyMiddleware(testAPIKey), nil)

	// when
	w := serve(t, r, map[string]string{"X-Organization-Id": testOrgID})

	// then
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_APIKeyMiddleware_shouldReturn403_whenAPIKeyIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	r := setupRouter(t, middleware.NewAPIKeyMiddleware(testAPIKey), nil)

	// when
	w := serve(t, r, map[string]string{"X-Service-Api-Key": testWrongKey, "X-Organization-Id": testOrgID})

	// then
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func Test_OrganizationIDHeaderMiddleware_shouldSetOrganizationID_whenHeaderIsPresent(t *testing.T) {
	t.Parallel()

	// given
	captured := &capturedContext{}
	r := setupRouter(t, middleware.NewOrganizationIDHeaderMiddleware(), captured)

	// when
	w := serve(t, r, map[string]string{"X-Organization-Id": testOrgID})

	// then
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, testOrgID, captured.OrganizationID)
}

func Test_OrganizationIDHeaderMiddleware_shouldReturn400_whenHeaderIsMissing(t *testing.T) {
	t.Parallel()

	// given
	r := setupRouter(t, middleware.NewOrganizationIDHeaderMiddleware(), nil)

	// when
	w := serve(t, r, nil)

	// then
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
