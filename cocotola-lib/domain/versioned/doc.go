// Package versioned provides shared primitives for aggregates and entities
// that use optimistic concurrency control via a monotonically increasing
// version field. Persistence helpers in cocotola-lib/gateway/gormsave and
// cocotola-lib/gateway/firestoresave consume types defined here.
package versioned
