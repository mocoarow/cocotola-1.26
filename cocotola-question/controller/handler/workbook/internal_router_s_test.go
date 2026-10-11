package workbook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// cocotola-init lists workbooks through the internal API for its idempotency check.
func Test_InitInternalWorkbookRouter_shouldRegisterListRoute_whenMounted(t *testing.T) {
	t.Parallel()

	// given
	r := initInternalWorkbookRouter(context.Background(), t, NewMockCreateWorkbookUsecase(t))

	// when
	routes := r.Routes()

	// then
	assert.Contains(t, routeKeys(routes), http.MethodGet+" /api/v1/internal/workbook")
}

func routeKeys(routes gin.RoutesInfo) []string {
	keys := make([]string, len(routes))
	for i, route := range routes {
		keys[i] = route.Method + " " + route.Path
	}

	return keys
}
