// Package idset provides a shared set data structure for managing IDs
// belonging to an owner entity. It is internal to the domain package and its subpackages.
package idset

// Set is a shared base type for aggregates that manage a set of IDs
// belonging to an owner entity. It is NOT safe for concurrent access;
// callers must ensure that a Set is only accessed from a single goroutine.
//
// The type parameter O is the owner ID type (e.g. OrganizationID, int).
// The type parameter M is the member ID type (e.g. AppUserID, int).
type Set[O any, M comparable] struct {
	// OwnerID is the ID of the entity that owns this set.
	OwnerID O
	entries map[M]struct{}
}

// New creates a new Set with the given owner ID and initial entries.
func New[O any, M comparable](ownerID O, ids []M) Set[O, M] {
	m := make(map[M]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}

	return Set[O, M]{OwnerID: ownerID, entries: m}
}

// Size returns the number of entries in the set.
func (s *Set[O, M]) Size() int { return len(s.entries) }

// Contains returns true if the given ID exists in the set.
func (s *Set[O, M]) Contains(id M) bool {
	_, ok := s.entries[id]
	return ok
}

// IDs returns a copy of the current IDs as a slice.
func (s *Set[O, M]) IDs() []M {
	result := make([]M, 0, len(s.entries))
	for id := range s.entries {
		result = append(result, id)
	}
	return result
}

// Add adds an ID to the set.
func (s *Set[O, M]) Add(id M) { s.entries[id] = struct{}{} }

// Remove removes an ID from the set.
func (s *Set[O, M]) Remove(id M) { delete(s.entries, id) }

// AddWithLimit adds an ID, checking for duplicates and capacity.
// Returns dupErr if the ID already exists, limitErr if the set is at capacity.
func (s *Set[O, M]) AddWithLimit(id M, limit int, limitErr error, dupErr error) error {
	if s.Contains(id) {
		return dupErr
	}
	if s.Size() >= limit {
		return limitErr
	}
	s.Add(id)
	return nil
}

// AddUnique adds an ID, checking only for duplicates.
// Returns dupErr if the ID already exists.
func (s *Set[O, M]) AddUnique(id M, dupErr error) error {
	if s.Contains(id) {
		return dupErr
	}
	s.Add(id)
	return nil
}
