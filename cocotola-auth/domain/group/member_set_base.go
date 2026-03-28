package group

import "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"

// memberSetBase provides common methods for group member set aggregates (users and child groups).
type memberSetBase struct{ set idset.Set }

// GroupID returns the group ID.
func (b *memberSetBase) GroupID() int { return b.set.OwnerID() }

// Size returns the number of entries in the set.
func (b *memberSetBase) Size() int { return b.set.Size() }

// Contains returns true if the given ID exists in the set.
func (b *memberSetBase) Contains(id int) bool { return b.set.Contains(id) }

// Remove removes an ID from the set.
func (b *memberSetBase) Remove(id int) { b.set.Remove(id) }
