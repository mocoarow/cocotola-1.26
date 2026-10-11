package versioned

import "errors"

// ErrConcurrentModification is returned when an optimistic-lock CAS fails but
// the row still exists; callers should reload and retry.
var ErrConcurrentModification = errors.New("concurrent modification")

// ErrNotFound is returned when a versioned update targets a deleted row;
// reloading cannot recover it.
var ErrNotFound = errors.New("versioned entity not found")
