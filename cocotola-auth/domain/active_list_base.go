package domain

import "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"

// activeListBase provides common methods for active user/group list aggregates.
type activeListBase struct{ set idset.Set }

// OrganizationID returns the organization ID.
func (b *activeListBase) OrganizationID() int { return b.set.OwnerID() }

// Entries returns a copy of the current IDs as a slice.
func (b *activeListBase) Entries() []int { return b.set.IDs() }

// Size returns the number of entries in the set.
func (b *activeListBase) Size() int { return b.set.Size() }

// Contains returns true if the given ID exists in the set.
func (b *activeListBase) Contains(id int) bool { return b.set.Contains(id) }

// Remove removes an ID from the set.
func (b *activeListBase) Remove(id int) { b.set.Remove(id) }
