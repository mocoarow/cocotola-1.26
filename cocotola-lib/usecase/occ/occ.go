package occ

import (
	"errors"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

// MaxAttempts is how many times Retry runs fn while it reports a concurrent modification.
const MaxAttempts = 5

// Retry runs fn again while it returns versioned.ErrConcurrentModification, up to MaxAttempts times.
func Retry(fn func() error) error {
	var err error
	for range MaxAttempts {
		err = fn()
		if !errors.Is(err, versioned.ErrConcurrentModification) {
			return err
		}
	}
	return err
}
