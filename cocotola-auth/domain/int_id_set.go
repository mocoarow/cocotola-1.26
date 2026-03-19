package domain

// intIDSet is a shared base type for aggregates that manage a set of integer IDs
// belonging to an owner entity.
type intIDSet struct {
	ownerID int
	entries map[int]struct{}
}

func newIntIDSet(ownerID int, ids []int) intIDSet {
	m := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return intIDSet{ownerID: ownerID, entries: m}
}

func (s *intIDSet) getOwnerID() int { return s.ownerID }

// Size returns the number of entries in the set.
func (s *intIDSet) Size() int { return len(s.entries) }

// Contains returns true if the given ID exists in the set.
func (s *intIDSet) Contains(id int) bool {
	_, ok := s.entries[id]
	return ok
}

func (s *intIDSet) ids() []int {
	result := make([]int, 0, len(s.entries))
	for id := range s.entries {
		result = append(result, id)
	}
	return result
}

func (s *intIDSet) add(id int) { s.entries[id] = struct{}{} }

// Remove removes an ID from the set.
func (s *intIDSet) Remove(id int) { delete(s.entries, id) }

// addWithLimit adds an ID, checking for duplicates and capacity.
func (s *intIDSet) addWithLimit(id int, limit int, limitErr error) error {
	if s.Contains(id) {
		return ErrDuplicateEntry
	}
	if s.Size() >= limit {
		return limitErr
	}
	s.add(id)
	return nil
}

// addUnique adds an ID, checking only for duplicates.
func (s *intIDSet) addUnique(id int) error {
	if s.Contains(id) {
		return ErrDuplicateEntry
	}
	s.add(id)
	return nil
}
