// Package firestoresave provides a generic Save helper that persists a
// versioned aggregate using Firestore with optimistic concurrency control.
//
// The helper runs a Firestore transaction that reads the current document,
// compares its version against the entity's expected version, and writes the
// new record only if they match. The entity's version is updated on success
// so callers do not have to manage versioning themselves.
package firestoresave
