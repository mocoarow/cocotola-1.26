package process_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/process"
)

func Test_Run_shouldReturnZero_whenAllProcessesSucceed(t *testing.T) {
	t.Parallel()

	// given
	successFunc := func(_ context.Context) process.RunProcess {
		return func() error {
			return nil
		}
	}

	// when
	code := process.Run(context.Background(), successFunc, successFunc)

	// then
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func Test_Run_shouldReturnZero_whenProcessReturnsCanceled(t *testing.T) {
	t.Parallel()

	// given
	cancelFunc := func(_ context.Context) process.RunProcess {
		return func() error {
			return context.Canceled
		}
	}

	// when
	code := process.Run(context.Background(), cancelFunc)

	// then
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func Test_Run_shouldReturnOne_whenProcessReturnsError(t *testing.T) {
	t.Parallel()

	// given
	errFunc := func(_ context.Context) process.RunProcess {
		return func() error {
			return errors.New("something went wrong")
		}
	}

	// when
	code := process.Run(context.Background(), errFunc)

	// then
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func Test_Run_shouldReturnOne_whenCanceledAndRealErrorAreMixed(t *testing.T) {
	t.Parallel()

	// given: synchronize both goroutines to trigger the race
	startCh := make(chan struct{})
	var readyWg sync.WaitGroup
	readyWg.Add(2)

	go func() {
		readyWg.Wait()
		close(startCh)
	}()

	cancelFunc := func(_ context.Context) process.RunProcess {
		return func() error {
			readyWg.Done()
			<-startCh
			return context.Canceled
		}
	}
	errFunc := func(_ context.Context) process.RunProcess {
		return func() error {
			readyWg.Done()
			<-startCh
			return errors.New("real error")
		}
	}

	// when
	code := process.Run(context.Background(), cancelFunc, errFunc)

	// then
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}

func Test_Run_shouldReturnOne_whenMultipleGoroutinesReturnErrorsConcurrently(t *testing.T) {
	t.Parallel()

	// given: start all three goroutines simultaneously to stress nonCanceledErr locking
	const numFuncs = 3
	startCh := make(chan struct{})
	var readyWg sync.WaitGroup
	readyWg.Add(numFuncs)

	go func() {
		readyWg.Wait()
		close(startCh)
	}()

	makeErrFunc := func(msg string) process.RunProcessFunc {
		return func(_ context.Context) process.RunProcess {
			return func() error {
				readyWg.Done()
				<-startCh
				return errors.New(msg)
			}
		}
	}

	// when
	code := process.Run(context.Background(),
		makeErrFunc("error1"),
		makeErrFunc("error2"),
		makeErrFunc("error3"),
	)

	// then
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
}
