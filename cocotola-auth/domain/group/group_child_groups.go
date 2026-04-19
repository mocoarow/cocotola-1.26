package group

import (
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"
)

// ChildGroups is an aggregate that manages the set of child group IDs belonging to a group.
type ChildGroups struct {
	memberSetBase[domain.GroupID]
}

// NewChildGroups creates a validated ChildGroups aggregate.
func NewChildGroups(groupID domain.GroupID, childGroupIDs []domain.GroupID) (*ChildGroups, error) {
	if groupID.IsZero() {
		return nil, fmt.Errorf("group child groups group id must not be zero: %w", domain.ErrInvalidArgument)
	}
	return &ChildGroups{memberSetBase[domain.GroupID]{idset.New[domain.GroupID, domain.GroupID](groupID, childGroupIDs)}}, nil
}

// ChildGroupIDs returns a copy of the current child group IDs as a slice.
func (g *ChildGroups) ChildGroupIDs() []domain.GroupID { return g.set.IDs() }

// Add adds a child group ID. Returns ErrDuplicateEntry if the group ID already exists.
func (g *ChildGroups) Add(childGroupID domain.GroupID) error {
	if err := g.set.AddUnique(childGroupID, domain.ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add child group: %w", err)
	}
	return nil
}
