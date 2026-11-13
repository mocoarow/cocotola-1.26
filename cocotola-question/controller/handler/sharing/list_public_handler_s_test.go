package sharing_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	r := initSharingRouter(ctx, t, NewMockListSharedUsecase(t), invalidLanguage)
	w := httptest.NewRecorder()

	// when
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/workbook/public", nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	respBytes := readBytes(t, w.Body)

	// then
	require.Equal(t, http.StatusBadRequest, w.Code)
	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(respBytes, &resp))
	assert.Equal(t, api.ErrorResponse{Code: "invalid_request", Message: "Bad Request"}, resp)
}
