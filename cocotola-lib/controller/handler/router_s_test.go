package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/handler"
)

func Test_InitRootRouterGroup_shouldReturnRouter_whenValidConfig(t *testing.T) {
	t.Parallel()

	// given
	config := controller.ServerConfig{
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
	config := controller.ServerConfig{
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

func panicRouter(t *testing.T, path string) http.Handler {
	t.Helper()

	config := controller.ServerConfig{
		CORS: controller.CORSConfig{
			AllowOrigins:     "*",
			AllowMethods:     "GET",
			AllowHeaders:     "",
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
	router, err := handler.InitRootRouterGroup(context.Background(), config, "test-app")
	require.NoError(t, err)
	router.GET(path, func(_ *gin.Context) {
		panic("boom")
	})

	return router
}

func serveGET(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, nil))

	return w
}

func Test_InitRootRouterGroup_shouldRespond500_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	path := "/router-test/respond-500"
	router := panicRouter(t, path)

	// when
	w := serveGET(t, router, path)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_InitRootRouterGroup_shouldRecordMetrics_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	path := "/router-test/record-metrics"
	router := panicRouter(t, path)

	// when
	serveGET(t, router, path)

	// then
	metrics := serveGET(t, promhttp.Handler(), "/metrics").Body.String()
	assert.Contains(t, metrics, `http_requests_total{method="GET",path="`+path+`",status="500"} 1`+"\n")
}

func Test_InitRootRouterGroup_shouldEnableAccessLog_whenAccessLogEnabled(t *testing.T) {
	t.Parallel()

	// given
	config := controller.ServerConfig{
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
