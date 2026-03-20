package gateway

import (
	"context"
	"fmt"
	"log/slog"

	libprocess "github.com/mocoarow/cocotola-1.26/cocotola-lib/process"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// EventHandler processes a domain event.
type EventHandler func(ctx context.Context, event domain.Event) error

// EventBus dispatches domain events to registered handlers via a buffered channel.
type EventBus struct {
	eventChan chan domain.Event
	handlers  map[string][]EventHandler
	logger    *slog.Logger
	started   bool
}

// NewEventBus returns a new EventBus with the given buffer size. bufferSize must be positive.
func NewEventBus(bufferSize int, logger *slog.Logger) *EventBus {
	if bufferSize <= 0 {
		panic(fmt.Sprintf("event_bus: bufferSize must be positive, got %d", bufferSize))
	}

	return &EventBus{
		eventChan: make(chan domain.Event, bufferSize),
		handlers:  make(map[string][]EventHandler),
		logger:    logger,
		started:   false,
	}
}

// Publish sends an event to the bus. If the buffer is full, the event is dropped with a log warning.
func (b *EventBus) Publish(event domain.Event) {
	select {
	case b.eventChan <- event:
	default:
		b.logger.Warn("event bus buffer full, dropping event",
			slog.String("event_type", event.EventType()))
	}
}

// Subscribe registers a handler for the given event type.
// Subscribe must be called before Start. Calling Subscribe after Start will panic.
func (b *EventBus) Subscribe(eventType string, handler EventHandler) {
	if b.started {
		panic("event_bus: Subscribe called after Start")
	}

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

// Start returns a RunProcess that dispatches events until ctx is canceled, then drains remaining events.
// Its signature satisfies libprocess.RunProcessFunc so it can be passed directly to libprocess.Run.
func (b *EventBus) Start(ctx context.Context) libprocess.RunProcess {
	b.started = true

	return func() error {
		b.logger.InfoContext(ctx, "event bus started")

		for {
			select {
			case event := <-b.eventChan:
				b.dispatch(ctx, event)
			case <-ctx.Done():
				b.drain()
				b.logger.InfoContext(ctx, "event bus stopped")

				return ctx.Err()
			}
		}
	}
}

func (b *EventBus) dispatch(ctx context.Context, event domain.Event) {
	handlers := b.handlers[event.EventType()]
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			b.logger.ErrorContext(ctx, "event handler failed",
				slog.String("event_type", event.EventType()),
				slog.Any("error", err))
		}
	}
}

func (b *EventBus) drain() {
	drainCtx := context.Background()

	for {
		select {
		case event := <-b.eventChan:
			b.dispatch(drainCtx, event)
		default:
			return
		}
	}
}
