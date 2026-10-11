package process

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// ErrGracefulShutdown signals that a process stopped because shutdown was requested (e.g. an OS signal).
var ErrGracefulShutdown = errors.New("graceful shutdown")

// RunProcess is a function that executes a process and returns an error on failure.
type RunProcess func() error

// RunProcessFunc creates a RunProcess bound to the given context.
type RunProcessFunc func(ctx context.Context) RunProcess

// Run executes all RunProcessFuncs concurrently and returns 0 on success or shutdown, or 1 on any other error.
func Run(ctx context.Context, runFuncs ...RunProcessFunc) int {
	var eg *errgroup.Group
	eg, ctx = errgroup.WithContext(ctx)

	errMu := &sync.Mutex{}
	var nonShutdownErr error

	for _, rf := range runFuncs {
		eg.Go(func() error {
			err := rf(ctx)()
			if err == nil {
				return nil
			}
			if !isShutdown(err) {
				errMu.Lock()
				if nonShutdownErr == nil {
					nonShutdownErr = err
				}
				errMu.Unlock()
			}

			return fmt.Errorf("run process: %w", err)
		})
	}

	if err := eg.Wait(); err != nil {
		if nonShutdownErr == nil && isShutdown(err) {
			return 0
		}

		return 1
	}

	return 0
}

func isShutdown(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, ErrGracefulShutdown)
}
