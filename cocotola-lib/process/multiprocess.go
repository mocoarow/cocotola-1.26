package process

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// RunProcess is a function that executes a process and returns an error on failure.
type RunProcess func() error

// RunProcessFunc creates a RunProcess bound to the given context.
type RunProcessFunc func(ctx context.Context) RunProcess

// Run executes all RunProcessFuncs concurrently and returns 0 on success or 1 on non-canceled error.
func Run(ctx context.Context, runFuncs ...RunProcessFunc) int {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	errMu := &sync.Mutex{}
	var nonCanceledErr error

	for _, rf := range runFuncs {
		eg.Go(func() error {
			err := rf(ctx)()
			if err == nil {
				return nil
			}
			if !errors.Is(err, context.Canceled) {
				errMu.Lock()
				if nonCanceledErr == nil {
					nonCanceledErr = err
				}
				errMu.Unlock()
			}

			return fmt.Errorf("run process: %w", err)
		})
	}

	if err := eg.Wait(); err != nil {
		if nonCanceledErr == nil && errors.Is(err, context.Canceled) {
			return 0
		}

		return 1
	}

	return 0
}
