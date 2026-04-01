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
