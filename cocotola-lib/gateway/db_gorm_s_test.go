package gateway_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

const testLoginID = "owner@example.com"

type loginRecord struct {
	LoginID string
}

type dryRunGormDBInput struct {
	logHandler     slog.Handler
	logLevel       slog.Level
	tracerProvider trace.TracerProvider
	queryErr       error
}

func openDryRunGormDB(t *testing.T, in dryRunGormDBInput) *gorm.DB {
	t.Helper()
	dialector := gormtests.DummyDialector{TranslatedErr: in.queryErr}
	db, err := gateway.OpenGormDBWithObservers(dialector, in.logHandler, in.logLevel, in.tracerProvider)
	require.NoError(t, err)
	if in.queryErr != nil {
		err := db.Callback().Query().After("gorm:query").Register("test:fail", func(tx *gorm.DB) {
			_ = tx.AddError(in.queryErr)
		})
		require.NoError(t, err)
	}
	return db.Session(&gorm.Session{DryRun: true})
}

func newBufferHandler(buf *bytes.Buffer) slog.Handler {
	return slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
}

func findByLoginID(db *gorm.DB) {
	var records []loginRecord
	db.Where("login_id = ?", testLoginID).Find(&records)
}

func Test_openGormDBWithObservers_shouldNotLogBoundValues_whenQueryFails(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		queryErr:       errors.New("duplicate key"),
	})

	// when
	findByLoginID(db)

	// then
	require.Contains(t, buf.String(), "login_id = ?")
	assert.NotContains(t, buf.String(), testLoginID)
}

func Test_openGormDBWithObservers_shouldNotLogBoundValues_whenSessionDebugIsEnabled(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		queryErr:       nil,
	})

	// when
	findByLoginID(db.Debug())

	// then
	require.Contains(t, buf.String(), "login_id = ?")
	assert.NotContains(t, buf.String(), testLoginID)
}

func Test_openGormDBWithObservers_shouldLogSuccessfulQuery_whenLogLevelIsDebug(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelDebug,
		tracerProvider: sdktrace.NewTracerProvider(),
		queryErr:       nil,
	})

	// when
	findByLoginID(db)

	// then
	assert.Contains(t, buf.String(), "login_id = ?")
}

func Test_openGormDBWithObservers_shouldNotLogSuccessfulQuery_whenLogLevelIsInfo(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		queryErr:       nil,
	})

	// when
	findByLoginID(db)

	// then
	assert.Empty(t, buf.String())
}

func Test_openGormDBWithObservers_shouldNotLog_whenRecordIsNotFound(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		queryErr:       gorm.ErrRecordNotFound,
	})

	// when
	findByLoginID(db)

	// then
	assert.Empty(t, buf.String())
}

func Test_openGormDBWithObservers_shouldNotRecordBoundValuesInSpanQueryText_whenQuerySucceeds(t *testing.T) {
	t.Parallel()

	// given
	recorder := tracetest.NewSpanRecorder()
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     slog.DiscardHandler,
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)),
		queryErr:       nil,
	})

	// when
	findByLoginID(db)

	// then
	spans := recorder.Ended()
	require.Len(t, spans, 1)
	var queryText string
	for _, attr := range spans[0].Attributes() {
		if attr.Key == semconv.DBQueryTextKey {
			queryText = attr.Value.AsString()
		}
	}
	require.Contains(t, queryText, "login_id = ?")
	assert.NotContains(t, queryText, testLoginID)
}
