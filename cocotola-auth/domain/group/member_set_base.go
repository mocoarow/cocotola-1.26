package group

import (
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"
)

// memberSetBase provides common methods for group member set aggregates (users and child groups).
type memberSetBase[M comparable] struct {
	set idset.Set[domain.GroupID, M]
}

// GroupID returns the group ID.
func (b *memberSetBase[M]) GroupID() domain.GroupID { return b.set.OwnerID }

// Size returns the number of entries in the set.
func (b *memberSetBase[M]) Size() int { return b.set.Size() }

// Contains returns true if the given ID exists in the set.
func (b *memberSetBase[M]) Contains(id M) bool { return b.set.Contains(id) }

// Remove removes an ID from the set.
func (b *memberSetBase[M]) Remove(id M) { b.set.Remove(id) }
