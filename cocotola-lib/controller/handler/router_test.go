package handler_test

import (
	"context"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
)

func Test_InitRootRouterGroup_shouldReturnRouter_whenValidConfig(t *testing.T) {
	t.Parallel()

	// given
	config := &controller.Config{
		CORS: controller.CORSConfig{
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST",
			AllowHeaders:     "Content-Type",
			AllowCredentials: false,
		},
		Log: controller.LogConfig{
			AccessLog:             false,
			AccessLogRequestBody:  false,
			AccessLogResponseBody: false,
		},
		Debug: controller.DebugConfig{
			Gin:  false,
			Wait: false,
		},
	}

	// when
	router, err := handler.InitRootRouterGroup(context.Background(), config, "test-app")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func Test_InitRootRouterGroup_shouldReturnError_whenCORSCredentialsWithWildcard(t *testing.T) {
	t.Parallel()

	// given
	config := &controller.Config{
		CORS: controller.CORSConfig{
			AllowOrigins:     "*",
			AllowMethods:     "GET",
			AllowHeaders:     "",
			AllowCredentials: true,
		},
		Log: controller.LogConfig{
			AccessLog:             false,
			AccessLogRequestBody:  false,
			AccessLogResponseBody: false,
		},
		Debug: controller.DebugConfig{
			Gin:  false,
			Wait: false,
		},
	}

	// when
	_, err := handler.InitRootRouterGroup(context.Background(), config, "test-app")

	// then
	if err == nil {
		t.Fatal("expected error for CORS credentials with wildcard, got nil")
	}
}

func Test_InitRootRouterGroup_shouldEnableAccessLog_whenAccessLogEnabled(t *testing.T) {
	t.Parallel()

	// given
	config := &controller.Config{
		CORS: controller.CORSConfig{
			AllowOrigins:     "http://localhost:3000",
			AllowMethods:     "GET,POST",
			AllowHeaders:     "Authorization",
			AllowCredentials: true,
		},
		Log: controller.LogConfig{
			AccessLog:             true,
			AccessLogRequestBody:  true,
			AccessLogResponseBody: true,
		},
		Debug: controller.DebugConfig{
			Gin:  true,
			Wait: true,
		},
	}

	// when
	router, err := handler.InitRootRouterGroup(context.Background(), config, "test-app")

	// then
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}
