package domain

import "errors"

// ErrWorkbookNotFound is returned when a workbook cannot be found.
var ErrWorkbookNotFound = errors.New("workbook not found")

// ErrQuestionNotFound is returned when a question cannot be found.
var ErrQuestionNotFound = errors.New("question not found")

// ErrReferenceNotFound is returned when a workbook reference cannot be found.
var ErrReferenceNotFound = errors.New("workbook reference not found")

// ErrDuplicateReference is returned when a user already references the same workbook.
var ErrDuplicateReference = errors.New("duplicate workbook reference")

// ErrForbidden is returned when the operator does not have permission to perform the action.
var ErrForbidden = errors.New("forbidden")

// ErrInvalidArgument is returned when a required field is missing, empty, or invalid.
var ErrInvalidArgument = errors.New("invalid argument")

// ErrOwnedWorkbookLimitReached is returned when the user has reached their owned workbook limit.
var ErrOwnedWorkbookLimitReached = errors.New("owned workbook limit reached")

// ErrDuplicateOwnedWorkbook is returned when attempting to add a workbook that is already owned.
var ErrDuplicateOwnedWorkbook = errors.New("duplicate owned workbook")

// ErrConcurrentModification is returned when an optimistic lock fails due to concurrent updates.
var ErrConcurrentModification = errors.New("concurrent modification")

// ErrStudyRecordNotFound is returned when a study record cannot be found.
var ErrStudyRecordNotFound = errors.New("study record not found")
