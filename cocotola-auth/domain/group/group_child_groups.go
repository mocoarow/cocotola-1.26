package group

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/internal/idset"
)

// ChildGroups is an aggregate that manages the set of child group IDs belonging to a group.
type ChildGroups struct{ memberSetBase }

// NewChildGroups creates a validated ChildGroups aggregate.
func NewChildGroups(groupID int, childGroupIDs []int) (*ChildGroups, error) {
	if groupID <= 0 {
		return nil, errors.New("group child groups group id must be positive")
	}
	return &ChildGroups{memberSetBase{idset.New(groupID, childGroupIDs)}}, nil
}

// ChildGroupIDs returns a copy of the current child group IDs as a slice.
func (g *ChildGroups) ChildGroupIDs() []int { return g.set.IDs() }

// Add adds a child group ID. Returns ErrDuplicateEntry if the group ID already exists.
func (g *ChildGroups) Add(childGroupID int) error {
	if err := g.set.AddUnique(childGroupID, domain.ErrDuplicateEntry); err != nil {
		return fmt.Errorf("add child group: %w", err)
	}
	return nil
}
