package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/middleware"
)

func Test_NewWaitMiddleware_shouldNotDelay_whenDurationIsZero(t *testing.T) {
	t.Parallel()

	// given
	gin.SetMode(gin.TestMode)
	mw := middleware.NewWaitMiddleware(0)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	// when
	start := time.Now()
	mw(c)
	elapsed := time.Since(start)

	// then
	if elapsed > 100*time.Millisecond {
		t.Fatalf("expected no delay, but took %v", elapsed)
	}
}

func Test_NewWaitMiddleware_shouldDelay_whenDurationIsPositive(t *testing.T) {
	t.Parallel()

	// given
	gin.SetMode(gin.TestMode)
	mw := middleware.NewWaitMiddleware(200 * time.Millisecond)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	// when
	start := time.Now()
	mw(c)
	elapsed := time.Since(start)

	// then
	if elapsed < 150*time.Millisecond {
		t.Fatalf("expected delay of ~200ms, but only took %v", elapsed)
	}
}

func Test_NewWaitMiddleware_shouldStopWaiting_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	gin.SetMode(gin.TestMode)
	mw := middleware.NewWaitMiddleware(5 * time.Second)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctx, cancel := context.WithCancel(context.Background())
	c.Request = httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)

	// when
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	mw(c)
	elapsed := time.Since(start)

	// then
	if elapsed > time.Second {
		t.Fatalf("expected cancellation to stop waiting, but took %v", elapsed)
	}
}

func Test_PrometheusMiddleware_shouldProcessRequest_whenCalled(t *testing.T) {
	t.Parallel()

	// given
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.PrometheusMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// when
	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	// then
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
