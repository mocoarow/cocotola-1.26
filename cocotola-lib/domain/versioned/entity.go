package versioned

// Entity is implemented by aggregates that use optimistic concurrency control.
// Version is 0 for a freshly constructed aggregate.
type Entity interface {
	Version() int
	SetVersion(int)
}

// Record is implemented by persistence records whose version is written by Save helpers.
type Record interface {
	GetVersion() int
}
