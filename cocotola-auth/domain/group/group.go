// Package group contains the group aggregate of the cocotola-auth domain.
package group

import (
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
		version:        0,
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
		version:        0,
		organizationID: organizationID,
		name:           name,
		enabled:        enabled,
	}
}

func (g *Group) validate() error {
	if g.id.IsZero() {
		return fmt.Errorf("group id must not be zero: %w", domain.ErrInvalidArgument)
	}
	if g.organizationID.IsZero() {
		return fmt.Errorf("group organization id must not be zero: %w", domain.ErrInvalidArgument)
	}
	if g.name == "" {
		return fmt.Errorf("group name is required: %w", domain.ErrInvalidArgument)
	}
	if len(g.name) > maxGroupNameLength {
		return fmt.Errorf("group name must not exceed %d characters: %w", maxGroupNameLength, domain.ErrInvalidArgument)
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

// Version returns the persisted row version (0 = new, not yet saved).
func (g *Group) Version() int { return g.version }

// SetVersion sets the persisted row version.
// Only the gateway/repository layer should call this, after a successful load or save.
func (g *Group) SetVersion(version int) {
	g.version = version
}
