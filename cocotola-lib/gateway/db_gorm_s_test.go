package gateway_test

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/codes"
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

type gormCallbackProcessor interface {
	Get(name string) func(*gorm.DB)
	Replace(name string, fn func(*gorm.DB)) error
}

type callbackFailure struct {
	processor func(db *gorm.DB) gormCallbackProcessor
	callback  string
	err       error
}

type dryRunGormDBInput struct {
	logHandler     slog.Handler
	logLevel       slog.Level
	tracerProvider trace.TracerProvider
	failure        *callbackFailure
}

func queryFailure(err error) *callbackFailure {
	return &callbackFailure{
		processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Query() },
		callback:  "gorm:query",
		err:       err,
	}
}

func openDryRunGormDB(t *testing.T, in dryRunGormDBInput) *gorm.DB {
	t.Helper()
	var translatedErr error
	if in.failure != nil {
		translatedErr = in.failure.err
	}
	dialector := gormtests.DummyDialector{TranslatedErr: translatedErr}
	db, err := gateway.OpenGormDBWithObservers(dialector, in.logHandler, in.logLevel, in.tracerProvider)
	require.NoError(t, err)
	if in.failure != nil {
		wrapBuiltinCallback(t, in.failure.processor(db), in.failure.callback, func(tx *gorm.DB) {
			_ = tx.AddError(in.failure.err)
		})
	}
	return db.Session(&gorm.Session{DryRun: true})
}

func wrapBuiltinCallback(t *testing.T, processor gormCallbackProcessor, name string, after func(tx *gorm.DB)) {
	t.Helper()
	builtin := processor.Get(name)
	require.NotNil(t, builtin)
	err := processor.Replace(name, func(tx *gorm.DB) {
		builtin(tx)
		after(tx)
	})
	require.NoError(t, err)
}

func newBufferHandler(buf *bytes.Buffer) slog.Handler {
	return slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})
}

func findByLoginID(db *gorm.DB) {
	var records []loginRecord
	db.Where("login_id = ?", testLoginID).Find(&records)
}

func newInvalidUUIDPgError() *pgconn.PgError {
	return &pgconn.PgError{
		Severity: "ERROR",
		Code:     "22P02",
		Message:  `invalid input syntax for type uuid: "` + testLoginID + `"`,
	}
}

func errorSpanStatusDescription(t *testing.T, recorder *tracetest.SpanRecorder) string {
	t.Helper()
	for _, span := range recorder.Ended() {
		if span.Status().Code == codes.Error {
			return span.Status().Description
		}
	}
	require.FailNow(t, "no span with error status")
	return ""
}

func Test_openGormDBWithObservers_shouldNotLogBoundValues_whenQueryFails(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		failure:        queryFailure(errors.New("duplicate key")),
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
		failure:        nil,
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
		failure:        nil,
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
		failure:        nil,
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
		failure:        queryFailure(gorm.ErrRecordNotFound),
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
		failure:        nil,
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

func Test_openGormDBWithObservers_shouldNotLogDriverErrorMessage_whenQueryFailsWithDriverError(t *testing.T) {
	t.Parallel()

	// given
	var buf bytes.Buffer
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     newBufferHandler(&buf),
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		failure:        queryFailure(newInvalidUUIDPgError()),
	})

	// when
	findByLoginID(db)

	// then
	require.Contains(t, buf.String(), "SQLSTATE 22P02")
	assert.NotContains(t, buf.String(), testLoginID)
}

type gormOperation struct {
	name      string
	processor func(db *gorm.DB) gormCallbackProcessor
	callback  string
	run       func(db *gorm.DB)
}

func transactionalGormOperations() []gormOperation {
	return []gormOperation{
		{
			name:      "create",
			processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Create() },
			callback:  "gorm:create",
			run:       func(db *gorm.DB) { db.Create(&loginRecord{LoginID: testLoginID}) },
		},
		{
			name:      "update",
			processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Update() },
			callback:  "gorm:update",
			run: func(db *gorm.DB) {
				db.Model(&loginRecord{}).Where("login_id = ?", testLoginID).Update("login_id", "renamed")
			},
		},
		{
			name:      "delete",
			processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Delete() },
			callback:  "gorm:delete",
			run:       func(db *gorm.DB) { db.Where("login_id = ?", testLoginID).Delete(&loginRecord{}) },
		},
	}
}

func Test_openGormDBWithObservers_shouldNotRecordDriverErrorMessageInSpanStatus_whenOperationFailsWithDriverError(t *testing.T) {
	t.Parallel()

	tests := append(transactionalGormOperations(), []gormOperation{
		{
			name:      "query",
			processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Query() },
			callback:  "gorm:query",
			run:       findByLoginID,
		},
		{
			name:      "row",
			processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Row() },
			callback:  "gorm:row",
			run:       func(db *gorm.DB) { db.Table("login_records").Where("login_id = ?", testLoginID).Row() },
		},
		{
			name:      "raw",
			processor: func(db *gorm.DB) gormCallbackProcessor { return db.Callback().Raw() },
			callback:  "gorm:raw",
			run:       func(db *gorm.DB) { db.Exec("DELETE FROM login_records WHERE login_id = ?", testLoginID) },
		},
	}...)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			recorder := tracetest.NewSpanRecorder()
			db := openDryRunGormDB(t, dryRunGormDBInput{
				logHandler:     slog.DiscardHandler,
				logLevel:       slog.LevelInfo,
				tracerProvider: sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)),
				failure: &callbackFailure{
					processor: tt.processor,
					callback:  tt.callback,
					err:       newInvalidUUIDPgError(),
				},
			})

			// when
			tt.run(db)

			// then
			description := errorSpanStatusDescription(t, recorder)
			require.Contains(t, description, "SQLSTATE 22P02")
			assert.NotContains(t, description, testLoginID)
		})
	}
}

func Test_openGormDBWithObservers_shouldNotRecordDriverErrorMessageInSpanStatus_whenRollbackAlsoFails(t *testing.T) {
	t.Parallel()

	for _, tt := range transactionalGormOperations() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			recorder := tracetest.NewSpanRecorder()
			db := openDryRunGormDB(t, dryRunGormDBInput{
				logHandler:     slog.DiscardHandler,
				logLevel:       slog.LevelInfo,
				tracerProvider: sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)),
				failure: &callbackFailure{
					processor: tt.processor,
					callback:  tt.callback,
					err:       newInvalidUUIDPgError(),
				},
			})
			wrapBuiltinCallback(t, tt.processor(db), "gorm:commit_or_rollback_transaction", func(tx *gorm.DB) {
				// DummyDialector would translate the rollback error into the injected driver error.
				tx.TranslateError = false
				_ = tx.AddError(errors.New("rollback failed"))
			})

			// when
			tt.run(db)

			// then
			description := errorSpanStatusDescription(t, recorder)
			require.Contains(t, description, "SQLSTATE 22P02")
			assert.NotContains(t, description, testLoginID)
		})
	}
}

func Test_openGormDBWithObservers_shouldNotRecordDriverErrorMessageInSpanStatus_whenTransactionEndFailsWithDriverError(t *testing.T) {
	t.Parallel()

	for _, tt := range transactionalGormOperations() {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			recorder := tracetest.NewSpanRecorder()
			db := openDryRunGormDB(t, dryRunGormDBInput{
				logHandler:     slog.DiscardHandler,
				logLevel:       slog.LevelInfo,
				tracerProvider: sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder)),
				failure: &callbackFailure{
					processor: tt.processor,
					callback:  "gorm:commit_or_rollback_transaction",
					err:       newInvalidUUIDPgError(),
				},
			})

			// when
			tt.run(db)

			// then
			description := errorSpanStatusDescription(t, recorder)
			require.Contains(t, description, "SQLSTATE 22P02")
			assert.NotContains(t, description, testLoginID)
		})
	}
}

func Test_openGormDBWithObservers_shouldReturnErrorWrappingDriverError_whenQueryFailsWithDriverError(t *testing.T) {
	t.Parallel()

	// given
	pgErr := newInvalidUUIDPgError()
	db := openDryRunGormDB(t, dryRunGormDBInput{
		logHandler:     slog.DiscardHandler,
		logLevel:       slog.LevelInfo,
		tracerProvider: sdktrace.NewTracerProvider(),
		failure:        queryFailure(pgErr),
	})

	// when
	var records []loginRecord
	err := db.Where("login_id = ?", testLoginID).Find(&records).Error

	// then
	assert.ErrorIs(t, err, pgErr)
}
