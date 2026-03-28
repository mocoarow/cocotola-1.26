// Package idset provides a shared set data structure for managing integer IDs
// belonging to an owner entity. It is internal to the domain package and its subpackages.
package idset

// Set is a shared base type for aggregates that manage a set of integer IDs
// belonging to an owner entity. It is NOT safe for concurrent access;
// callers must ensure that a Set is only accessed from a single goroutine.
type Set struct {
	ownerID int
	entries map[int]struct{}
}

// New creates a new Set with the given owner ID and initial entries.
// Precondition: ownerID must be positive. Callers are expected to validate this.
func New(ownerID int, ids []int) Set {
	m := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return Set{ownerID: ownerID, entries: m}
}

// OwnerID returns the owner ID.
func (s *Set) OwnerID() int { return s.ownerID }

// Size returns the number of entries in the set.
func (s *Set) Size() int { return len(s.entries) }

// Contains returns true if the given ID exists in the set.
func (s *Set) Contains(id int) bool {
	_, ok := s.entries[id]
	return ok
}

// IDs returns a copy of the current IDs as a slice.
func (s *Set) IDs() []int {
	result := make([]int, 0, len(s.entries))
	for id := range s.entries {
		result = append(result, id)
	}
	return result
}

// Add adds an ID to the set.
func (s *Set) Add(id int) { s.entries[id] = struct{}{} }

// Remove removes an ID from the set.
func (s *Set) Remove(id int) { delete(s.entries, id) }

// AddWithLimit adds an ID, checking for duplicates and capacity.
// Returns dupErr if the ID already exists, limitErr if the set is at capacity.
func (s *Set) AddWithLimit(id int, limit int, limitErr error, dupErr error) error {
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
func (s *Set) AddUnique(id int, dupErr error) error {
	if s.Contains(id) {
		return dupErr
	}
	s.Add(id)
	return nil
}
