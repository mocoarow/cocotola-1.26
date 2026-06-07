package gateway_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

type captureHandler struct {
	handled []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.handled = append(h.handled, r)
	return nil
}

func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(_ string) slog.Handler { return h }

func Test_stringToLogLevel_shouldReturnExpectedLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"uppercase DEBUG", "DEBUG", slog.LevelDebug},
		{"unsupported level", "unknown", slog.LevelWarn},
		{"empty string", "", slog.LevelWarn},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			level := gateway.StringToLogLevel(tt.input)

			// then
			if level != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, level)
			}
		})
	}
}

func Test_levelFilterHandler_Enabled_shouldReturnFalse_whenLevelBelowMinLevel(t *testing.T) {
	t.Parallel()

	// given
	h := gateway.NewLevelFilterHandlerForTest(&captureHandler{}, slog.LevelWarn)

	// when
	result := h.Enabled(context.Background(), slog.LevelInfo)

	// then
	assert.False(t, result)
}

func Test_levelFilterHandler_Enabled_shouldReturnTrue_whenLevelAtOrAboveMinLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level slog.Level
	}{
		{"at minLevel", slog.LevelWarn},
		{"above minLevel", slog.LevelError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			h := gateway.NewLevelFilterHandlerForTest(&captureHandler{}, slog.LevelWarn)

			// when
			result := h.Enabled(context.Background(), tt.level)

			// then
			assert.True(t, result)
		})
	}
}

func Test_levelFilterHandler_Handle_shouldBlockRecord_whenLevelBelowMinLevel(t *testing.T) {
	t.Parallel()

	// given
	inner := &captureHandler{}
	h := gateway.NewLevelFilterHandlerForTest(inner, slog.LevelWarn)
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	// when
	err := h.Handle(context.Background(), record)

	// then
	require.NoError(t, err)
	assert.Empty(t, inner.handled)
}

func Test_levelFilterHandler_Handle_shouldDelegateRecord_whenLevelAtOrAboveMinLevel(t *testing.T) {
	t.Parallel()

	// given
	inner := &captureHandler{}
	h := gateway.NewLevelFilterHandlerForTest(inner, slog.LevelWarn)
	record := slog.NewRecord(time.Now(), slog.LevelWarn, "test message", 0)

	// when
	err := h.Handle(context.Background(), record)

	// then
	require.NoError(t, err)
	assert.Len(t, inner.handled, 1)
}

func Test_levelFilterHandler_WithAttrs_shouldPreserveMinLevel(t *testing.T) {
	t.Parallel()

	// given
	h := gateway.NewLevelFilterHandlerForTest(&captureHandler{}, slog.LevelWarn)

	// when
	h2 := h.WithAttrs([]slog.Attr{slog.String("key", "value")})

	// then
	assert.False(t, h2.Enabled(context.Background(), slog.LevelInfo), "level below minLevel should be disabled")
	assert.True(t, h2.Enabled(context.Background(), slog.LevelWarn), "level at minLevel should be enabled")
}

func Test_levelFilterHandler_WithGroup_shouldPreserveMinLevel(t *testing.T) {
	t.Parallel()

	// given
	h := gateway.NewLevelFilterHandlerForTest(&captureHandler{}, slog.LevelWarn)

	// when
	h2 := h.WithGroup("group")

	// then
	assert.False(t, h2.Enabled(context.Background(), slog.LevelInfo), "level below minLevel should be disabled")
	assert.True(t, h2.Enabled(context.Background(), slog.LevelWarn), "level at minLevel should be enabled")
}
