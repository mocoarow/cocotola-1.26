package telemetry_test

import (
	"context"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/baggage"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/telemetry"
)

func Test_AddBaggageMembers_shouldAddMember_whenSingleValueProvided(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	logger := slog.Default()
	values := map[string]string{"key1": "value1"}

	// when
	ctx = telemetry.AddBaggageMembers(ctx, values, logger)

	// then
	bag := baggage.FromContext(ctx)
	member := bag.Member("key1")
	if member.Value() != "value1" {
		t.Fatalf("expected baggage member key1=value1, got %s", member.Value())
	}
}

func Test_AddBaggageMembers_shouldAddAllMembers_whenMultipleValuesProvided(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	logger := slog.Default()
	values := map[string]string{"k1": "v1", "k2": "v2"}

	// when
	ctx = telemetry.AddBaggageMembers(ctx, values, logger)

	// then
	bag := baggage.FromContext(ctx)
	if bag.Member("k1").Value() != "v1" {
		t.Fatalf("expected k1=v1, got %s", bag.Member("k1").Value())
	}
	if bag.Member("k2").Value() != "v2" {
		t.Fatalf("expected k2=v2, got %s", bag.Member("k2").Value())
	}
}

func Test_AddBaggageMembers_shouldSkipInvalidKey_whenKeyContainsInvalidChars(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	logger := slog.Default()
	values := map[string]string{
		"valid-key":   "value1",
		"invalid key": "value2",
	}

	// when
	ctx = telemetry.AddBaggageMembers(ctx, values, logger)

	// then
	bag := baggage.FromContext(ctx)
	if bag.Member("valid-key").Value() != "value1" {
		t.Fatalf("expected valid-key=value1, got %s", bag.Member("valid-key").Value())
	}
	if bag.Member("invalid key").Value() != "" {
		t.Fatalf("expected invalid key to be skipped, got %s", bag.Member("invalid key").Value())
	}
}

func Test_AddBaggageMembers_shouldReturnOriginalContext_whenEmptyValues(t *testing.T) {
	t.Parallel()

	// given
	ctx := context.Background()
	logger := slog.Default()
	values := map[string]string{}

	// when
	newCtx := telemetry.AddBaggageMembers(ctx, values, logger)

	// then
	bag := baggage.FromContext(newCtx)
	if len(bag.Members()) != 0 {
		t.Fatalf("expected no baggage members, got %d", len(bag.Members()))
	}
}
