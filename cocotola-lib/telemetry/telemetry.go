package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/baggage"
)

// AddBaggageMembers adds the given key-value pairs as OpenTelemetry baggage members to the context.
func AddBaggageMembers(ctx context.Context, values map[string]string, logger *slog.Logger) context.Context {
	bag := baggage.FromContext(ctx)
	for key, value := range values {
		member, err := baggage.NewMember(key, value)
		if err != nil {
			logger.WarnContext(ctx, "new baggage member", slog.Any("error", err), slog.String("key", key))

			continue
		}

		newBag, err := bag.SetMember(member)
		if err != nil {
			logger.WarnContext(ctx, "set baggage member", slog.Any("error", err), slog.String("key", key))

			continue
		}

		bag = newBag
	}

	ctx = baggage.ContextWithBaggage(ctx, bag)
	return ctx
}
