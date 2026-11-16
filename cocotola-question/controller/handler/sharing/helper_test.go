package sharing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	sharinghandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/sharing"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	fixtureUserID         = "user-1"
	fixtureOrganizationID = "org-1"
)

var serverConfig libcontroller.ServerConfig

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	serverConfig = libcontroller.ServerConfig{
		CORS: libcontroller.CORSConfig{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Content-Type,Authorization,X-Service-Api-Key",
		},
		Log: libcontroller.LogConfig{
			AccessLog:             false,
			AccessLogRequestBody:  false,
			AccessLogResponseBody: false,
		},
		Debug: libcontroller.DebugConfig{
			Gin:  false,
			Wait: false,
		},
	}
	os.Exit(m.Run())
}

func fakeAuthMiddleware(userID string, organizationID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, userID)
		c.Set(controller.ContextFieldOrganizationID{}, organizationID)
		c.Next()
	}
}

type sharingUsecases struct {
	share      *MockShareWorkbookUsecase
	listShared *MockListSharedUsecase
	unshare    *MockUnshareUsecase
	listPublic *MockListPublicUsecase
}

func newSharingUsecases(t *testing.T) sharingUsecases {
	t.Helper()

	return sharingUsecases{
		share:      NewMockShareWorkbookUsecase(t),
		listShared: NewMockListSharedUsecase(t),
		unshare:    NewMockUnshareUsecase(t),
		listPublic: NewMockListPublicUsecase(t),
	}
}

func initSharingRouter(ctx context.Context, t *testing.T, usecases sharingUsecases, authMiddleware gin.HandlerFunc, middleware ...gin.HandlerFunc) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)
	parent := router.Group("api").Group("v1")

	sharinghandler.InitSharingRouter(
		sharinghandler.NewShareWorkbookHandler(usecases.share),
		sharinghandler.NewListSharedHandler(usecases.listShared),
		sharinghandler.NewUnshareHandler(usecases.unshare),
		sharinghandler.NewListPublicHandler(usecases.listPublic),
		parent,
		authMiddleware,
		middleware...,
	)

	return router
}

func serve(ctx context.Context, t *testing.T, r *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(ctx, method, path, nil)
	require.NoError(t, err)
	r.ServeHTTP(w, req)
	return w
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}

func readErrorResponse(t *testing.T, w *httptest.ResponseRecorder) api.ErrorResponse {
	t.Helper()

	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(readBytes(t, w.Body), &resp))
	return resp
}
