// Package group contains the group aggregate of the cocotola-auth domain.
package group

import (
	"errors"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

const maxGroupNameLength = 255

// Group represents a group belonging to an organization.
type Group struct {
	id             domain.GroupID
	version        int
	organizationID domain.OrganizationID
	name           string
	enabled        bool
}

// NewGroup creates a validated Group.
func NewGroup(id domain.GroupID, organizationID domain.OrganizationID, name string, enabled bool) (*Group, error) {
	m := &Group{
		id:             id,
		version:        1,
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
func ReconstructGroup(id domain.GroupID, organizationID domain.OrganizationID, name string, enabled bool) *Group {
	return &Group{
		id:             id,
		organizationID: organizationID,
		name:           name,
		enabled:        enabled,
	}
}

func (g *Group) validate() error {
	if g.id.IsZero() {
		return errors.New("group id must not be zero")
	}
	if g.organizationID.IsZero() {
		return errors.New("group organization id must not be zero")
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
func (g *Group) ID() domain.GroupID { return g.id }

// OrganizationID returns the organization ID.
func (g *Group) OrganizationID() domain.OrganizationID { return g.organizationID }

// Name returns the group name.
func (g *Group) Name() string { return g.name }

// Enabled returns whether the group is enabled.
func (g *Group) Enabled() bool { return g.enabled }

// Enable enables the group.
func (g *Group) Enable() { g.enabled = true }

// Disable disables the group.
func (g *Group) Disable() { g.enabled = false }

// Version returns the persisted row version (1 = new, not yet saved).
func (g *Group) Version() int { return g.version }

// IncrementVersion bumps the version after a successful persist.
func (g *Group) IncrementVersion() { g.version++ }

// WithVersion sets the persisted row version on a reconstituted aggregate.
func (g *Group) WithVersion(version int) *Group {
	g.version = version
	return g
}
