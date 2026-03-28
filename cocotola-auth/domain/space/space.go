package space

import (
	"errors"
	"fmt"
)

const (
	maxSpaceKeyNameLength = 50
	maxSpaceNameLength    = 100
)

// Space represents a space belonging to an organization.
type Space struct {
	id             int
	organizationID int
	ownerID        int
	keyName        string
	name           string
	spaceType      Type
	deleted        bool
}

// NewSpace creates a validated Space.
func NewSpace(id int, organizationID int, ownerID int, keyName string, name string, spaceType Type, deleted bool) (*Space, error) {
	m := &Space{
		id:             id,
		organizationID: organizationID,
		ownerID:        ownerID,
		keyName:        keyName,
		name:           name,
		spaceType:      spaceType,
		deleted:        deleted,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructSpace reconstitutes a Space from persistence.
func ReconstructSpace(id int, organizationID int, ownerID int, keyName string, name string, spaceType Type, deleted bool) *Space {
	return &Space{
		id:             id,
		organizationID: organizationID,
		ownerID:        ownerID,
		keyName:        keyName,
		name:           name,
		spaceType:      spaceType,
		deleted:        deleted,
	}
}

func (s *Space) validate() error {
	if s.id <= 0 {
		return errors.New("space id must be positive")
	}
	if s.organizationID <= 0 {
		return errors.New("space organization id must be positive")
	}
	if s.ownerID <= 0 {
		return errors.New("space owner id must be positive")
	}
	if s.keyName == "" {
		return errors.New("space key name is required")
	}
	if len(s.keyName) > maxSpaceKeyNameLength {
		return fmt.Errorf("space key name must not exceed %d characters", maxSpaceKeyNameLength)
	}
	if s.name == "" {
		return errors.New("space name is required")
	}
	if len(s.name) > maxSpaceNameLength {
		return fmt.Errorf("space name must not exceed %d characters", maxSpaceNameLength)
	}
	if s.spaceType.Value() == "" {
		return errors.New("space type is required")
	}
	return nil
}

// ID returns the space ID.
func (s *Space) ID() int { return s.id }

// OrganizationID returns the organization ID.
func (s *Space) OrganizationID() int { return s.organizationID }

// OwnerID returns the owner user ID.
func (s *Space) OwnerID() int { return s.ownerID }

// KeyName returns the space key name.
func (s *Space) KeyName() string { return s.keyName }

// Name returns the space name.
func (s *Space) Name() string { return s.name }

// SpaceType returns the space type.
func (s *Space) SpaceType() Type { return s.spaceType }

// Deleted returns whether the space is soft-deleted.
func (s *Space) Deleted() bool { return s.deleted }

// Delete marks the space as deleted.
func (s *Space) Delete() { s.deleted = true }

// Restore marks the space as not deleted.
func (s *Space) Restore() { s.deleted = false }

// PublicSpaceKeyName returns the key name for a public space in the given organization.
func PublicSpaceKeyName(orgName string) string {
	return "public@@" + orgName
}

// PrivateSpaceKeyName returns the key name for a private space belonging to the given user.
func PrivateSpaceKeyName(loginID string) string {
	return "private@@" + loginID
}
