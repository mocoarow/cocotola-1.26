package domain

import (
	"errors"
	"fmt"
)

const maxGroupNameLength = 255

// Group represents a group belonging to an organization.
type Group struct {
	id             int
	organizationID int
	name           string
	enabled        bool
}

// NewGroup creates a validated Group.
func NewGroup(id int, organizationID int, name string, enabled bool) (*Group, error) {
	m := &Group{
		id:             id,
		organizationID: organizationID,
		name:           name,
		enabled:        enabled,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructGroup reconstitutes a Group from persistence.
func ReconstructGroup(id int, organizationID int, name string, enabled bool) *Group {
	return &Group{
		id:             id,
		organizationID: organizationID,
		name:           name,
		enabled:        enabled,
	}
}

func (g *Group) validate() error {
	if g.id <= 0 {
		return errors.New("group id must be positive")
	}
	if g.organizationID <= 0 {
		return errors.New("group organization id must be positive")
	}
	if g.name == "" {
		return errors.New("group name is required")
	}
	if len(g.name) > maxGroupNameLength {
		return fmt.Errorf("group name must not exceed %d characters", maxGroupNameLength)
	}
	return nil
}

// ID returns the group ID.
func (g *Group) ID() int { return g.id }

// OrganizationID returns the organization ID.
func (g *Group) OrganizationID() int { return g.organizationID }

// Name returns the group name.
func (g *Group) Name() string { return g.name }

// Enabled returns whether the group is enabled.
func (g *Group) Enabled() bool { return g.enabled }

// Enable enables the group.
func (g *Group) Enable() { g.enabled = true }

// Disable disables the group.
func (g *Group) Disable() { g.enabled = false }
