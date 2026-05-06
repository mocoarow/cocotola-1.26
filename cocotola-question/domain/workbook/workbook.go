package workbook

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
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
	language       Language
	version        int
	createdAt      time.Time
	updatedAt      time.Time
}

// NewWorkbook creates a validated Workbook with version=0 (a new aggregate not yet saved).
// Callers (usecase layer) must supply the ID and timestamps.
func NewWorkbook(id string, spaceID string, ownerID string, organizationID string, title string, description string, visibility Visibility, language Language, createdAt time.Time, updatedAt time.Time) (*Workbook, error) {
	m := &Workbook{
		id:             id,
		spaceID:        spaceID,
		ownerID:        ownerID,
		organizationID: organizationID,
		title:          title,
		description:    description,
		visibility:     visibility,
		language:       language,
		version:        0,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("new workbook: %w", err)
	}
	return m, nil
}

// ReconstructWorkbook reconstitutes a Workbook from persistence without validation.
func ReconstructWorkbook(id string, spaceID string, ownerID string, organizationID string, title string, description string, visibility Visibility, language Language, version int, createdAt time.Time, updatedAt time.Time) *Workbook {
	return &Workbook{
		id:             id,
		spaceID:        spaceID,
		ownerID:        ownerID,
		organizationID: organizationID,
		title:          title,
		description:    description,
		visibility:     visibility,
		language:       language,
		version:        version,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (w *Workbook) validate() error {
	if w.id == "" {
		return fmt.Errorf("workbook id is required: %w", domain.ErrInvalidArgument)
	}
	if w.spaceID == "" {
		return fmt.Errorf("workbook space id is required: %w", domain.ErrInvalidArgument)
	}
	if w.ownerID == "" {
		return fmt.Errorf("workbook owner id is required: %w", domain.ErrInvalidArgument)
	}
	if w.organizationID == "" {
		return fmt.Errorf("workbook organization id is required: %w", domain.ErrInvalidArgument)
	}
	if w.title == "" {
		return fmt.Errorf("workbook title is required: %w", domain.ErrInvalidArgument)
	}
	if len(w.title) > maxTitleLength {
		return fmt.Errorf("workbook title must not exceed %d characters: %w", maxTitleLength, domain.ErrInvalidArgument)
	}
	if len(w.description) > maxDescriptionLength {
		return fmt.Errorf("workbook description must not exceed %d characters: %w", maxDescriptionLength, domain.ErrInvalidArgument)
	}
	if w.visibility.Value() == "" {
		return fmt.Errorf("workbook visibility is required: %w", domain.ErrInvalidArgument)
	}
	if w.language.IsZero() {
		return fmt.Errorf("workbook language is required: %w", domain.ErrInvalidArgument)
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

// Language returns the workbook language.
func (w *Workbook) Language() Language { return w.language }

// Version returns the persisted version (0 = new, not yet saved).
func (w *Workbook) Version() int { return w.version }

// SetVersion sets the persisted version on the aggregate.
// Intended for repository implementations to update the version after a successful save.
// Do not call from application or domain code.
func (w *Workbook) SetVersion(version int) { w.version = version }

// CreatedAt returns the creation timestamp.
func (w *Workbook) CreatedAt() time.Time { return w.createdAt }

// UpdatedAt returns the last update timestamp.
func (w *Workbook) UpdatedAt() time.Time { return w.updatedAt }

// ChangeVisibility changes the workbook's visibility setting.
func (w *Workbook) ChangeVisibility(v Visibility) {
	w.visibility = v
}

// ChangeLanguage changes the workbook's language setting. Callers must pass a
// non-zero Language constructed via NewLanguage; the type system already
// guarantees the new value has been validated.
func (w *Workbook) ChangeLanguage(l Language) {
	w.language = l
}

// UpdateTitle updates the workbook title.
func (w *Workbook) UpdateTitle(title string) error {
	if title == "" {
		return fmt.Errorf("workbook title is required: %w", domain.ErrInvalidArgument)
	}
	if len(title) > maxTitleLength {
		return fmt.Errorf("workbook title must not exceed %d characters: %w", maxTitleLength, domain.ErrInvalidArgument)
	}
	w.title = title
	return nil
}

// UpdateDescription updates the workbook description.
func (w *Workbook) UpdateDescription(desc string) error {
	if len(desc) > maxDescriptionLength {
		return fmt.Errorf("workbook description must not exceed %d characters: %w", maxDescriptionLength, domain.ErrInvalidArgument)
	}
	w.description = desc
	return nil
}

// Touch updates the last-modified timestamp. Callers (usecase layer) invoke
// this before persisting so that the stored updatedAt reflects the time of
// the edit rather than the time the aggregate was loaded.
func (w *Workbook) Touch(now time.Time) {
	w.updatedAt = now
}
