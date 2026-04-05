package auth_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/controller"
	authhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/auth"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var testCookieConfig = controller.CookieConfig{
	Name:     "session_token",
	Path:     "/",
	Secure:   false,
	SameSite: "Lax",
}

// noopMiddleware is a pass-through middleware for tests that don't require authentication context.
func noopMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// fakeAuthMiddleware sets the given userID and loginID into the Gin context,
// simulating what the real auth middleware does.
func fakeAuthMiddleware(userID int, loginID string, organizationName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, userID)
		c.Set(controller.ContextFieldLoginID{}, loginID)
		c.Set(controller.ContextFieldOrganizationName{}, organizationName)
		c.Next()
	}
}

func initAuthRouter(ctx context.Context, t *testing.T, usecase *MockAuthUsecase) *gin.Engine {
	t.Helper()
	return initAuthRouterWithMiddleware(ctx, t, usecase, noopMiddleware())
}

func initAuthRouterWithMiddleware(ctx context.Context, t *testing.T, usecase *MockAuthUsecase, authMiddleware gin.HandlerFunc) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, config, domain.AppName)
	require.NoError(t, err)
	api := router.Group("api")
	v1 := api.Group("v1")

	authenticateHandler := authhandler.NewPasswordAuthenticateHandler(usecase, testCookieConfig, 30)
	guestAuthenticateHandler := authhandler.NewGuestAuthenticateHandler(usecase)
	refreshHandler := authhandler.NewRefreshHandler(usecase)
	revokeHandler := authhandler.NewRevokeHandler(usecase, testCookieConfig)
	getMeHandler := authhandler.NewGetMeHandler()
	authhandler.InitAuthRouter(authenticateHandler, guestAuthenticateHandler, refreshHandler, revokeHandler, getMeHandler, v1, authMiddleware)

	// internal routes (no API key middleware in tests)
	internalV1 := api.Group("v1/internal")
	supabaseExchangeHandler := authhandler.NewSupabaseExchangeHandler(usecase)
	authhandler.InitInternalAuthRouter(supabaseExchangeHandler, internalV1)

	return router
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}

func parseJSON(t *testing.T, bytes []byte) any {
	t.Helper()
	obj, err := oj.Parse(bytes)
	require.NoError(t, err)
	return obj
}

func parseExpr(t *testing.T, v string) jp.Expr {
	t.Helper()
	expr, err := jp.ParseString(v)
	require.NoError(t, err)
	return expr
}

func validateErrorResponse(t *testing.T, respBytes []byte, expectedErrorCode string, expectedErrorMessage string) {
	t.Helper()

	jsonObj := parseJSON(t, respBytes)

	// - error code
	errorCodeExpr := parseExpr(t, "$.code")
	errorCode := errorCodeExpr.Get(jsonObj)
	require.Len(t, errorCode, 1, "response should have one code: %+v", jsonObj)
	assert.Equal(t, expectedErrorCode, errorCode[0])

	// - error message
	errorMessageExpr := parseExpr(t, "$.message")
	errorMessage := errorMessageExpr.Get(jsonObj)
	require.Len(t, errorMessage, 1, "response should have one message: %+v", jsonObj)
	assert.Equal(t, expectedErrorMessage, errorMessage[0])
}

func findCookieByName(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()
	for _, c := range cookies {
		if c.Name == name {
			return c
		}
	}
	return nil
}
