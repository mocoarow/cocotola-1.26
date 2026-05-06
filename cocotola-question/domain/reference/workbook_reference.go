// Package reference contains the workbook reference aggregate of the cocotola-question domain.
package reference

import (
	"fmt"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

// WorkbookReference represents a user's reference link to another workbook.
// A reference is uniquely identified by (userID, workbookID); the persistence
// layer keys the document by workbookID under the user's subcollection so the
// uniqueness invariant is enforced by Firestore.Create rather than a TOCTOU
// query.
type WorkbookReference struct {
	userID     string
	workbookID string
	addedAt    time.Time
}

// NewWorkbookReference creates a validated WorkbookReference.
func NewWorkbookReference(userID string, workbookID string, addedAt time.Time) (*WorkbookReference, error) {
	m := &WorkbookReference{
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
func ReconstructWorkbookReference(userID string, workbookID string, addedAt time.Time) *WorkbookReference {
	return &WorkbookReference{
		userID:     userID,
		workbookID: workbookID,
		addedAt:    addedAt,
	}
}

func (r *WorkbookReference) validate() error {
	if r.userID == "" {
		return fmt.Errorf("workbook reference user id is required: %w", domain.ErrInvalidArgument)
	}
	if r.workbookID == "" {
		return fmt.Errorf("workbook reference workbook id is required: %w", domain.ErrInvalidArgument)
	}
	return nil
}

// ID returns the reference identifier. Because a reference is uniquely
// identified by (userID, workbookID) and persisted under a doc keyed by
// workbookID, ID is the same value as WorkbookID and exists only to satisfy
// callers that treat the reference as an opaque resource.
func (r *WorkbookReference) ID() string { return r.workbookID }

// UserID returns the referencing user ID.
func (r *WorkbookReference) UserID() string { return r.userID }

// WorkbookID returns the referenced workbook ID.
func (r *WorkbookReference) WorkbookID() string { return r.workbookID }

// AddedAt returns when the reference was added.
func (r *WorkbookReference) AddedAt() time.Time { return r.addedAt }
