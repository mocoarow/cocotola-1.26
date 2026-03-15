package gateway_test

import (
	"log/slog"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway"
)

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
