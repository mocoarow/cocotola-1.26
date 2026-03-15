package process_test

import (
	"context"
	"errors"
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
