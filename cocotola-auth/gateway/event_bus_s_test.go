package gateway_test

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/gateway"
)

func Test_EventBus_shouldDispatchEvents_whenHandlersRegistered(t *testing.T) {
	t.Parallel()

	// given
	logger := slog.Default()
	bus := gateway.NewEventBus(10, logger)

	var mu sync.Mutex
	var received []domain.Event

	bus.Subscribe(domain.EventTypeAppUserCreated, func(_ context.Context, event domain.Event) error {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, event)
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	runProcess := bus.Start(ctx)

	var wg sync.WaitGroup

	wg.Go(func() {
		_ = runProcess()
	})

	// when
	now := time.Now()
	event := domain.NewAppUserCreated(42, 1, "user@example.com", now)
	bus.Publish(event)

	// then
	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(received) == 1
	}, time.Second, 10*time.Millisecond)

	mu.Lock()
	got, ok := received[0].(domain.AppUserCreated)
	mu.Unlock()
	require.True(t, ok)
	assert.Equal(t, 42, got.AppUserID())
	assert.Equal(t, 1, got.OrganizationID())

	cancel()
	wg.Wait()
}

func Test_EventBus_shouldLogError_whenHandlerFails(t *testing.T) {
	t.Parallel()

	// given
	logger := slog.Default()
	bus := gateway.NewEventBus(10, logger)

	handlerErr := errors.New("handler failure")
	bus.Subscribe(domain.EventTypeAppUserCreated, func(_ context.Context, _ domain.Event) error {
		return handlerErr
	})

	var secondHandlerCalled sync.WaitGroup
	secondHandlerCalled.Add(1)
	bus.Subscribe(domain.EventTypeAppUserCreated, func(_ context.Context, _ domain.Event) error {
		secondHandlerCalled.Done()
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	runProcess := bus.Start(ctx)

	var wg sync.WaitGroup

	wg.Go(func() {
		_ = runProcess()
	})

	// when
	bus.Publish(domain.NewAppUserCreated(1, 1, "test", time.Now()))

	// then - second handler still called despite first handler error
	secondHandlerCalled.Wait()

	cancel()
	wg.Wait()
}

func Test_EventBus_shouldDrainEvents_whenContextCanceled(t *testing.T) {
	t.Parallel()

	// given
	logger := slog.Default()
	bus := gateway.NewEventBus(10, logger)

	var mu sync.Mutex
	var count int

	bus.Subscribe(domain.EventTypeAppUserCreated, func(_ context.Context, _ domain.Event) error {
		mu.Lock()
		defer mu.Unlock()

		count++

		return nil
	})

	// publish events before starting
	bus.Publish(domain.NewAppUserCreated(1, 1, "a", time.Now()))
	bus.Publish(domain.NewAppUserCreated(2, 1, "b", time.Now()))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// when
	runProcess := bus.Start(ctx)
	_ = runProcess()

	// then - events in buffer should be drained
	mu.Lock()
	assert.Equal(t, 2, count)
	mu.Unlock()
}
