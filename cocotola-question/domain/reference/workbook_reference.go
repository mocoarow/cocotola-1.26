// Package reference contains the workbook reference aggregate of the cocotola-question domain.
package reference

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// WorkbookReference represents a user's reference link to another workbook.
type WorkbookReference struct {
	id         string
	userID     string
	workbookID string
	addedAt    time.Time
}

// NewWorkbookReference creates a validated WorkbookReference.
func NewWorkbookReference(id string, userID string, workbookID string, addedAt time.Time) (*WorkbookReference, error) {
	m := &WorkbookReference{
		id:         id,
		userID:     userID,
		workbookID: workbookID,
		addedAt:    addedAt,
	}
	if err := m.validate(); err != nil {
		return nil, fmt.Errorf("new workbook reference: %w", err)
	}
	return m, nil
}

// ReconstructWorkbookReference reconstitutes a WorkbookReference from persistence without validation.
func ReconstructWorkbookReference(id string, userID string, workbookID string, addedAt time.Time) *WorkbookReference {
	return &WorkbookReference{
		id:         id,
		userID:     userID,
		workbookID: workbookID,
		addedAt:    addedAt,
	}
}

func (r *WorkbookReference) validate() error {
	if r.id == "" {
		return fmt.Errorf("workbook reference id is required: %w", domain.ErrInvalidArgument)
	}
	if r.userID == "" {
		return fmt.Errorf("workbook reference user id is required: %w", domain.ErrInvalidArgument)
	}
	if r.workbookID == "" {
		return fmt.Errorf("workbook reference workbook id is required: %w", domain.ErrInvalidArgument)
	}
	return nil
}

// ID returns the reference ID.
func (r *WorkbookReference) ID() string { return r.id }

// UserID returns the referencing user ID.
func (r *WorkbookReference) UserID() string { return r.userID }

// WorkbookID returns the referenced workbook ID.
func (r *WorkbookReference) WorkbookID() string { return r.workbookID }

// AddedAt returns when the reference was added.
func (r *WorkbookReference) AddedAt() time.Time { return r.addedAt }
