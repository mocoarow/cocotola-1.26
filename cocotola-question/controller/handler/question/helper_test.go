package question_test

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

	questionhandler "github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler/question"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	fixtureWorkbookID = "wb-1"
	fixtureQuestionID = "q-1"
	// fixtureAudioInputHash is a 64-char hex sha256 used by the audio batch.
	fixtureAudioInputHash = "a1b2c3d4e5f60718293a4b5c6d7e8f9001020304050607080910111213141516"
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

func initInternalAudioRouter(ctx context.Context, t *testing.T, usecase *MockAudioBatchUsecase) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(ctx, serverConfig, domain.AppName)
	require.NoError(t, err)
	api := router.Group("api")
	v1 := api.Group("v1")
	internal := v1.Group("internal")

	handler := questionhandler.NewAudioHandler(usecase)
	questionhandler.InitInternalAudioRouter(handler, internal)

	return router
}

func readBytes(t *testing.T, b *bytes.Buffer) []byte {
	t.Helper()
	respBytes, err := io.ReadAll(b)
	require.NoError(t, err)
	return respBytes
}

func parseJSON(t *testing.T, data []byte) any {
	t.Helper()
	obj, err := oj.Parse(data)
	require.NoError(t, err)
	return obj
}

func parseExpr(t *testing.T, v string) jp.Expr {
	t.Helper()
	expr, err := jp.ParseString(v)
	require.NoError(t, err)
	return expr
}

func validateErrorCode(t *testing.T, respBytes []byte, expectedCode string) {
	t.Helper()
	jsonObj := parseJSON(t, respBytes)
	codeExpr := parseExpr(t, "$.code")
	code := codeExpr.Get(jsonObj)
	require.Len(t, code, 1, "response should have one code: %+v", jsonObj)
	assert.Equal(t, expectedCode, code[0])
}
