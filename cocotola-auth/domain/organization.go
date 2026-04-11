package domain

import "fmt"

const maxOrganizationNameLength = 255

// Organization represents a tenant that owns users and groups.
type Organization struct {
	id              OrganizationID
	version         int
	name            string
	maxActiveUsers  int
	maxActiveGroups int
}

// NewOrganization creates a validated Organization.
func NewOrganization(id OrganizationID, name string, maxActiveUsers int, maxActiveGroups int) (*Organization, error) {
	m := &Organization{
		id:              id,
		version:         0,
		name:            name,
		maxActiveUsers:  maxActiveUsers,
		maxActiveGroups: maxActiveGroups,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructOrganization reconstitutes an Organization from persistence.
func ReconstructOrganization(id OrganizationID, name string, maxActiveUsers int, maxActiveGroups int) *Organization {
	return &Organization{
		id:              id,
		version:         0,
		name:            name,
		maxActiveUsers:  maxActiveUsers,
		maxActiveGroups: maxActiveGroups,
	}
}

func (o *Organization) validate() error {
	if o.id.IsZero() {
		return fmt.Errorf("organization id must not be zero: %w", ErrInvalidArgument)
	}
	if o.name == "" {
		return fmt.Errorf("organization name is required: %w", ErrInvalidArgument)
	}
	if len(o.name) > maxOrganizationNameLength {
		return fmt.Errorf("organization name must not exceed %d characters: %w", maxOrganizationNameLength, ErrInvalidArgument)
	}
	if o.maxActiveUsers <= 0 {
		return fmt.Errorf("organization max active users must be positive: %w", ErrInvalidArgument)
	}
	if o.maxActiveGroups <= 0 {
		return fmt.Errorf("organization max active groups must be positive: %w", ErrInvalidArgument)
	}
	return nil
}

// ID returns the organization ID.
func (o *Organization) ID() OrganizationID { return o.id }

// Name returns the organization name.
func (o *Organization) Name() string { return o.name }

// MaxActiveUsers returns the maximum number of active users.
func (o *Organization) MaxActiveUsers() int { return o.maxActiveUsers }

// MaxActiveGroups returns the maximum number of active groups.
func (o *Organization) MaxActiveGroups() int { return o.maxActiveGroups }

// Version returns the persisted row version (0 = new, not yet saved).
func (o *Organization) Version() int { return o.version }

// WithVersion sets the persisted row version on a reconstituted aggregate.
func (o *Organization) WithVersion(version int) *Organization {
	o.version = version
	return o
}
