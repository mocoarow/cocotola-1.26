package domain

import (
	"fmt"
	"slices"
	"sort"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/idset"
)

// ActiveQuestionList is an aggregate that manages the set of active question IDs
// within a workbook. It enables efficient lookup of which questions are available
// for study without loading full question data.
type ActiveQuestionList struct {
	set     idset.Set[string, string]
	version int
}

// NewActiveQuestionList creates a validated ActiveQuestionList.
func NewActiveQuestionList(workbookID string, questionIDs []string) (*ActiveQuestionList, error) {
	if workbookID == "" {
		return nil, fmt.Errorf("active question list workbook id is required: %w", ErrInvalidArgument)
	}
	if slices.Contains(questionIDs, "") {
		return nil, fmt.Errorf("question id is required: %w", ErrInvalidArgument)
	}
	return &ActiveQuestionList{set: idset.New(workbookID, questionIDs), version: 0}, nil
}

// Add adds a question ID to the list.
func (l *ActiveQuestionList) Add(questionID string) error {
	if questionID == "" {
		return fmt.Errorf("question id is required: %w", ErrInvalidArgument)
	}
	l.set.Add(questionID)
	return nil
}

// Remove removes a question ID from the list.
func (l *ActiveQuestionList) Remove(questionID string) { l.set.Remove(questionID) }

// WorkbookID returns the workbook ID.
func (l *ActiveQuestionList) WorkbookID() string { return l.set.OwnerID }

// Version returns the persisted version (0 = new, not yet saved).
func (l *ActiveQuestionList) Version() int { return l.version }

// SetVersion sets the persisted version on a reconstituted aggregate.
func (l *ActiveQuestionList) SetVersion(version int) { l.version = version }

// Entries returns a sorted copy of the current question IDs as a slice.
func (l *ActiveQuestionList) Entries() []string {
	result := l.set.IDs()
	sort.Strings(result)
	return result
}

// Size returns the number of entries in the set.
func (l *ActiveQuestionList) Size() int { return l.set.Size() }

// Contains returns true if the given question ID exists in the set.
func (l *ActiveQuestionList) Contains(questionID string) bool { return l.set.Contains(questionID) }
