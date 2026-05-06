package versioned

// Entity is implemented by aggregates and entities that use optimistic
// concurrency control. Version returns the version that was loaded from
// storage (0 for a freshly constructed aggregate); SetVersion is called by
// the persistence layer to update it after a successful save.
type Entity interface {
	Version() int
	SetVersion(int)
}

// Record is implemented by persistence record types whose Version field is
// written to storage by Save helpers. It exposes the version as a method so
// generic helpers can verify it without reflection. Implementations typically
// return the value of a Version int field on the record.
type Record interface {
	GetVersion() int
}
