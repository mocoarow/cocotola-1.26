package domain

import "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"

// activeListBase provides common methods for active user/group list aggregates.
// O is the owner ID type (OrganizationID), M is the member ID type.
type activeListBase[M comparable] struct {
	set idset.Set[OrganizationID, M]
}

// OrganizationID returns the organization ID.
func (b *activeListBase[M]) OrganizationID() OrganizationID { return b.set.OwnerID }

// Entries returns a copy of the current IDs as a slice.
func (b *activeListBase[M]) Entries() []M { return b.set.IDs() }

// Size returns the number of entries in the set.
func (b *activeListBase[M]) Size() int { return b.set.Size() }

// Contains returns true if the given ID exists in the set.
func (b *activeListBase[M]) Contains(id M) bool { return b.set.Contains(id) }

// Remove removes an ID from the set.
func (b *activeListBase[M]) Remove(id M) { b.set.Remove(id) }
