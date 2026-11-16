package sharing_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
)

func Test_ListPublicHandler_ListPublic_shouldReturn400_whenUserLanguageIsInvalid(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	invalidLanguage := func(c *gin.Context) {
		c.Set(controller.ContextFieldUserLanguage{}, "jpn")
		c.Next()
	}
	r := initSharingRouter(ctx, t, newSharingUsecases(t), fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID), invalidLanguage)

	// when
	w := serve(ctx, t, r, http.MethodGet, "/api/v1/workbook/public")

	// then
	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, api.ErrorResponse{Code: "invalid_request", Message: "Bad Request"}, readErrorResponse(t, w))
}
