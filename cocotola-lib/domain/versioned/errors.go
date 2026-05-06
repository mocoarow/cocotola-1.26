package versioned

import "errors"

// ErrConcurrentModification is returned when an optimistic-lock compare-and-swap
// fails because another transaction modified the aggregate between when it was
// loaded and when the save was attempted. The row identified by the primary key
// still exists; callers should reload the aggregate and retry.
var ErrConcurrentModification = errors.New("concurrent modification")

// ErrNotFound is returned by Save helpers when a versioned update targets a row
// that no longer exists (e.g., it was deleted between load and save). Callers
// cannot recover by reloading; they must surface this as a not-found condition.
var ErrNotFound = errors.New("versioned entity not found")
