package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/controller/middleware"
)

const panicPath = "/panic"

func newPanicRouter(t *testing.T, logger *slog.Logger, tp *sdktrace.TracerProvider) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(otelgin.Middleware("test", otelgin.WithTracerProvider(tp)))
	router.Use(middleware.NewRecoveryMiddleware(logger))
	router.GET(panicPath, func(_ *gin.Context) {
		panic("boom")
	})

	return router
}

func newRecorderTracerProvider(t *testing.T) (*sdktrace.TracerProvider, *tracetest.SpanRecorder) {
	t.Helper()

	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))

	return tp, recorder
}

func Test_NewRecoveryMiddleware_shouldRespond500_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	tp, _ := newRecorderTracerProvider(t)
	router := newPanicRouter(t, slog.New(slog.DiscardHandler), tp)

	// when
	w := serveGET(t, router, panicPath)

	// then
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func Test_NewRecoveryMiddleware_shouldSetSpanStatusError_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	tp, recorder := newRecorderTracerProvider(t)
	router := newPanicRouter(t, slog.New(slog.DiscardHandler), tp)

	// when
	serveGET(t, router, panicPath)

	// then
	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, codes.Error, spans[0].Status().Code)
}

func Test_NewRecoveryMiddleware_shouldDescribePanicInSpanStatus_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	tp, recorder := newRecorderTracerProvider(t)
	router := newPanicRouter(t, slog.New(slog.DiscardHandler), tp)

	// when
	serveGET(t, router, panicPath)

	// then
	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Contains(t, spans[0].Status().Description, "panic: boom")
}

func Test_NewRecoveryMiddleware_shouldRecordPanicOnSpan_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	tp, recorder := newRecorderTracerProvider(t)
	router := newPanicRouter(t, slog.New(slog.DiscardHandler), tp)

	// when
	serveGET(t, router, panicPath)

	// then
	spans := recorder.Ended()
	require.Len(t, spans, 1)
	events := spans[0].Events()
	require.Len(t, events, 1)
	assert.Equal(t, "exception", events[0].Name)
}

func Test_NewRecoveryMiddleware_shouldLogErrorWithStack_whenHandlerPanics(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	tp, _ := newRecorderTracerProvider(t)
	router := newPanicRouter(t, logger, tp)

	// when
	serveGET(t, router, panicPath)

	// then
	assert.Contains(t, buf.String(), `"level":"ERROR"`)
	assert.Contains(t, buf.String(), "boom")
	assert.Contains(t, buf.String(), `"stack":"goroutine`)
}

func Test_NewRecoveryMiddleware_shouldPassThrough_whenHandlerSucceeds(t *testing.T) {
	t.Parallel()

	// given
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.NewRecoveryMiddleware(slog.New(slog.DiscardHandler)))
	router.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	// when
	w := serveGET(t, router, "/ok")

	// then
	assert.Equal(t, http.StatusNoContent, w.Code)
}
