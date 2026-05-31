package authz_test

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	authzhandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/authz"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// routeKey identifies a registered gin route by method and path.
type routeKey struct {
	method string
	path   string
}

func registeredRoutes(t *testing.T, register func(internalAuthV1 gin.IRouter)) map[routeKey]struct{} {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(context.Background(), serverConfig, domain.AppName)
	require.NoError(t, err)
	internalAuthV1 := router.Group("api").Group("v1").Group("internal").Group("auth")

	register(internalAuthV1)

	routes := make(map[routeKey]struct{})
	for _, r := range router.Routes() {
		routes[routeKey{method: r.Method, path: r.Path}] = struct{}{}
	}
	return routes
}

// Test_InitInternalAuthzRouter_shouldRegisterRoute guards against the regression
// where the standalone (main.go) wiring registered the authz check route
// internally but omitted the policy route, making cocotola-question's
// per-resource policy grants 404 in production.
func Test_InitInternalAuthzRouter_shouldRegisterRoute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want routeKey
	}{
		{
			name: "policy route",
			want: routeKey{method: "POST", path: "/api/v1/internal/auth/authz/policy"},
		},
		{
			name: "check route",
			want: routeKey{method: "POST", path: "/api/v1/internal/auth/authz/check"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			checkHandler := authzhandler.NewCheckHandler(NewMockAuthorizationChecker(t))
			addPolicyHandler := authzhandler.NewAddPolicyHandler(NewMockUserPolicyAdder(t))

			// when
			routes := registeredRoutes(t, func(internalAuthV1 gin.IRouter) {
				authzhandler.InitInternalAuthzRouter(checkHandler, addPolicyHandler, internalAuthV1)
			})

			// then
			_, ok := routes[tt.want]
			assert.True(t, ok, "%s %s must be registered", tt.want.method, tt.want.path)
		})
	}
}
