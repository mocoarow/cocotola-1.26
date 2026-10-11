package question_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// cocotola-init lists questions through the internal API for its idempotency check.
func Test_InitInternalQuestionRouter_shouldRegisterListRoute_whenMounted(t *testing.T) {
	t.Parallel()

	// given
	r := initInternalQuestionRouter(context.Background(), t, NewMockAddQuestionUsecase(t), NewMockUpdateQuestionUsecase(t))

	// when
	routes := r.Routes()

	// then
	assert.Contains(t, routeKeys(routes), http.MethodGet+" /api/v1/internal/workbook/:workbookId/question")
}

func routeKeys(routes gin.RoutesInfo) []string {
	keys := make([]string, len(routes))
	for i, route := range routes {
		keys[i] = route.Method + " " + route.Path
	}

	return keys
}
