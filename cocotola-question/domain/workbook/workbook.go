package workbook

import (
	"errors"
	"fmt"
	"time"
)

const (
	maxTitleLength       = 200
	maxDescriptionLength = 1000
)

// Workbook is the aggregate root for the workbook aggregate.
type Workbook struct {
	id             string
	spaceID        string
	ownerID        string
	organizationID string
	title          string
	description    string
	visibility     Visibility
	createdAt      time.Time
	updatedAt      time.Time
}

// NewWorkbook creates a validated Workbook.
func NewWorkbook(id string, spaceID string, ownerID string, organizationID string, title string, description string, visibility Visibility, createdAt time.Time, updatedAt time.Time) (*Workbook, error) {
	m := &Workbook{
		id:             id,
		spaceID:        spaceID,
		ownerID:        ownerID,
		organizationID: organizationID,
		title:          title,
		description:    description,
		visibility:     visibility,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
	if err := m.validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// ReconstructWorkbook reconstitutes a Workbook from persistence without validation.
func ReconstructWorkbook(id string, spaceID string, ownerID string, organizationID string, title string, description string, visibility Visibility, createdAt time.Time, updatedAt time.Time) *Workbook {
	return &Workbook{
		id:             id,
		spaceID:        spaceID,
		ownerID:        ownerID,
		organizationID: organizationID,
		title:          title,
		description:    description,
		visibility:     visibility,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (w *Workbook) validate() error {
	if w.id == "" {
		return errors.New("workbook id is required")
	}
	if w.spaceID == "" {
		return errors.New("workbook space id is required")
	}
	if w.ownerID == "" {
		return errors.New("workbook owner id is required")
	}
	if w.organizationID == "" {
		return errors.New("workbook organization id is required")
	}
	if w.title == "" {
		return errors.New("workbook title is required")
	}
	if len(w.title) > maxTitleLength {
		return fmt.Errorf("workbook title must not exceed %d characters", maxTitleLength)
	}
	if len(w.description) > maxDescriptionLength {
		return fmt.Errorf("workbook description must not exceed %d characters", maxDescriptionLength)
	}
	if w.visibility.Value() == "" {
		return errors.New("workbook visibility is required")
	}
	return nil
}

// ID returns the workbook ID.
func (w *Workbook) ID() string { return w.id }

// SpaceID returns the space ID.
func (w *Workbook) SpaceID() string { return w.spaceID }

// OwnerID returns the owner user ID.
func (w *Workbook) OwnerID() string { return w.ownerID }

// OrganizationID returns the organization ID.
func (w *Workbook) OrganizationID() string { return w.organizationID }

// Title returns the workbook title.
func (w *Workbook) Title() string { return w.title }

// Description returns the workbook description.
func (w *Workbook) Description() string { return w.description }

// Visibility returns the visibility setting.
func (w *Workbook) Visibility() Visibility { return w.visibility }

// CreatedAt returns the creation timestamp.
func (w *Workbook) CreatedAt() time.Time { return w.createdAt }

// UpdatedAt returns the last update timestamp.
func (w *Workbook) UpdatedAt() time.Time { return w.updatedAt }

// ChangeVisibility changes the workbook's visibility setting.
func (w *Workbook) ChangeVisibility(v Visibility) {
	w.visibility = v
}

// UpdateTitle updates the workbook title.
func (w *Workbook) UpdateTitle(title string) error {
	if title == "" {
		return errors.New("workbook title is required")
	}
	if len(title) > maxTitleLength {
		return fmt.Errorf("workbook title must not exceed %d characters", maxTitleLength)
	}
	w.title = title
	return nil
}

// UpdateDescription updates the workbook description.
func (w *Workbook) UpdateDescription(desc string) error {
	if len(desc) > maxDescriptionLength {
		return fmt.Errorf("workbook description must not exceed %d characters", maxDescriptionLength)
	}
	w.description = desc
	return nil
}
