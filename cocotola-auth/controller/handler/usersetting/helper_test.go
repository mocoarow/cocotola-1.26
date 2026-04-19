package usersetting_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/stretchr/testify/require"

	libhandler "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"

	usersettinghandler "github.com/mocoarow/cocotola-1.26/cocotola-auth/controller/handler/usersetting"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

var (
	fixtureAppUserID = domain.MustParseAppUserID("00000000-0000-7000-8000-000000000020")
)

func initUserSettingRouter(_ context.Context, t *testing.T, settingFinder *mockuserSettingFinder) *gin.Engine {
	t.Helper()

	router, err := libhandler.InitRootRouterGroup(context.Background(), serverConfig, domain.AppName)
	require.NoError(t, err)
	api := router.Group("api")
	v1 := api.Group("v1")
	internalV1 := v1.Group("internal/auth")

	findHandler := usersettinghandler.NewFindUserSettingHandler(settingFinder)
	usersettinghandler.InitUserSettingRouter(findHandler, internalV1)

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

func validateErrorResponse(t *testing.T, respBytes []byte, expectedErrorCode string, expectedErrorMessage string) {
	t.Helper()

	jsonObj := parseJSON(t, respBytes)

	errorCodeExpr := parseExpr(t, "$.code")
	errorCode := errorCodeExpr.Get(jsonObj)
	require.Len(t, errorCode, 1)
	require.Equal(t, expectedErrorCode, errorCode[0])

	errorMessageExpr := parseExpr(t, "$.message")
	errorMessage := errorMessageExpr.Get(jsonObj)
	require.Len(t, errorMessage, 1)
	require.Equal(t, expectedErrorMessage, errorMessage[0])
}
