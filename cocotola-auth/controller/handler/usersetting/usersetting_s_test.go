package usersetting_test

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

var serverConfig libcontroller.ServerConfig

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	serverConfig = libcontroller.ServerConfig{
		CORS: libcontroller.CORSConfig{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Content-Type,Authorization",
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
