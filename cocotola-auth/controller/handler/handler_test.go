package handler_test

import (
	"os"
	"testing"

	libcontroller "github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
)

var (
	config libcontroller.Config
)

func TestMain(m *testing.M) {
	config = libcontroller.Config{
		CORS: libcontroller.CORSConfig{
			AllowOrigins: "*",
			AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders: "Content-Type",
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

	code := m.Run()

	os.Exit(code)
}
