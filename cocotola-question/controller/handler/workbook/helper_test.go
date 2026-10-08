package workbook_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller"
	workbookhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/workbook"
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

func fakeOperatorMiddleware(userID string, organizationID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(controller.ContextFieldUserID{}, userID)
		c.Set(controller.ContextFieldOrganizationID{}, organizationID)
		c.Next()
	}
}

func initInternalWorkbookRouter(ctx context.Context, t *testing.T, createUsecase *MockCreateWorkbookUsecase) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)
	internal := router.Group("api").Group("v1").Group("internal")
	internal.Use(fakeOperatorMiddleware(fixtureUserID, fixtureOrganizationID))

	workbookhandler.InitInternalWorkbookRouter(
		workbookhandler.NewCreateWorkbookHandler(createUsecase),
		workbookhandler.NewListWorkbooksHandler(NewMockListWorkbooksUsecase(t)),
		workbookhandler.NewUpdateWorkbookHandler(NewMockUpdateWorkbookUsecase(t)),
		workbookhandler.NewDeleteWorkbookHandler(NewMockDeleteWorkbookUsecase(t)),
		internal,
	)

	return router
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}

func validateErrorResponse(t *testing.T, respBytes []byte, expectedErrorCode string, expectedErrorMessage string) {
	t.Helper()

	jsonObj, err := oj.Parse(respBytes)
	require.NoError(t, err)

	codeExpr, err := jp.ParseString("$.code")
	require.NoError(t, err)
	code := codeExpr.Get(jsonObj)
	require.Len(t, code, 1, "response should have one code: %+v", jsonObj)
	assert.Equal(t, expectedErrorCode, code[0])

	messageExpr, err := jp.ParseString("$.message")
	require.NoError(t, err)
	message := messageExpr.Get(jsonObj)
	require.Len(t, message, 1, "response should have one message: %+v", jsonObj)
	assert.Equal(t, expectedErrorMessage, message[0])
}
