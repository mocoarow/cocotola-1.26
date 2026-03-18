package domain

import (
	"errors"
	"fmt"
)

const maxOrganizationNameLength = 255

// Organization represents a tenant that owns users and groups.
type Organization struct {
	id              int
	name            string
	maxActiveUsers  int
	maxActiveGroups int
}

// NewOrganization creates a validated Organization.
func NewOrganization(id int, name string, maxActiveUsers int, maxActiveGroups int) (*Organization, error) {
	m := &Organization{
		id:              id,
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
func ReconstructOrganization(id int, name string, maxActiveUsers int, maxActiveGroups int) *Organization {
	return &Organization{
		id:              id,
		name:            name,
		maxActiveUsers:  maxActiveUsers,
		maxActiveGroups: maxActiveGroups,
	}
}

func (o *Organization) validate() error {
	if o.id <= 0 {
		return errors.New("organization id must be positive")
	}
	if o.name == "" {
		return errors.New("organization name is required")
	}
	if len(o.name) > maxOrganizationNameLength {
		return fmt.Errorf("organization name must not exceed %d characters", maxOrganizationNameLength)
	}
	if o.maxActiveUsers <= 0 {
		return errors.New("organization max active users must be positive")
	}
	if o.maxActiveGroups <= 0 {
		return errors.New("organization max active groups must be positive")
	}
	return nil
}

// ID returns the organization ID.
func (o *Organization) ID() int { return o.id }

// Name returns the organization name.
func (o *Organization) Name() string { return o.name }

// MaxActiveUsers returns the maximum number of active users.
func (o *Organization) MaxActiveUsers() int { return o.maxActiveUsers }

// MaxActiveGroups returns the maximum number of active groups.
func (o *Organization) MaxActiveGroups() int { return o.maxActiveGroups }
