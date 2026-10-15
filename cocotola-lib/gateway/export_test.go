package gateway

import (
	"context"
	"log/slog"
	"time"
)

// MySQL defaults exported for testing.
const (
	MySQLMaxAllowedPacket = mysqlMaxAllowedPacket
	MySQLDefaultCharset   = mysqlDefaultCharset
	MySQLDefaultCollation = mysqlDefaultCollation
)

// StringToLogLevel exports stringToLogLevel for testing.
var StringToLogLevel = stringToLogLevel

// InitTraceSampler exports initTraceSampler for testing.
var InitTraceSampler = initTraceSampler

// NewLevelFilterHandlerForTest creates a levelFilterHandler for testing.
func NewLevelFilterHandlerForTest(handler slog.Handler, minLevel slog.Level) slog.Handler {
	return &levelFilterHandler{
		handler:  handler,
		minLevel: minLevel,
	}
}

// WatchSignal exports watchSignal for testing.
var WatchSignal = watchSignal

// NewShutdownCtxForTest exports newShutdownCtx for testing.
func NewShutdownCtxForTest(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return newShutdownCtx(ctx, timeout)
}

// OpenGormDBWithObservers exports openGormDBWithObservers for testing.
var OpenGormDBWithObservers = openGormDBWithObservers
