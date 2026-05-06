// Package gormsave provides a generic Save helper that persists a versioned
// aggregate using GORM with optimistic concurrency control.
//
// Aggregates with version 0 are inserted; aggregates with version > 0 are
// updated via a compare-and-swap on the version column. The helper updates
// the aggregate's version on success so callers do not have to manage
// versioning themselves.
package gormsave
