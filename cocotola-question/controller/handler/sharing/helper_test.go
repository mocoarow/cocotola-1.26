package sharing_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

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

func initSharingRouter(ctx context.Context, t *testing.T, listSharedUsecase *MockListSharedUsecase) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)
	parent := router.Group("api").Group("v1")

	sharinghandler.InitSharingRouter(
		sharinghandler.NewShareWorkbookHandler(NewMockShareWorkbookUsecase(t)),
		sharinghandler.NewListSharedHandler(listSharedUsecase),
		sharinghandler.NewUnshareHandler(NewMockUnshareUsecase(t)),
		sharinghandler.NewListPublicHandler(NewMockListPublicUsecase(t)),
		parent,
		fakeAuthMiddleware(fixtureUserID, fixtureOrganizationID),
	)

	return router
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}
