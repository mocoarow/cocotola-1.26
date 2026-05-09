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

// ErrStudyRecordNotFound is returned when a study record cannot be found.
var ErrStudyRecordNotFound = errors.New("study record not found")

// ErrActiveQuestionListNotFound is returned when an active question list cannot be found.
var ErrActiveQuestionListNotFound = errors.New("active question list not found")

// ErrOwnedWorkbookListNotFound is returned when an owned workbook list cannot be found.
var ErrOwnedWorkbookListNotFound = errors.New("owned workbook list not found")

// ErrSpaceNotFound is returned when a space lookup against cocotola-auth returns 404.
var ErrSpaceNotFound = errors.New("space not found")

// ErrInvalidSpaceType is returned when a space lookup against cocotola-auth
// returns a SpaceType value that this service does not recognize.
var ErrInvalidSpaceType = errors.New("invalid space type")

// ErrAudioInputHashMismatch is returned by audio batch state transitions when
// the supplied input hash does not match the question's current audio
// generation hash (typically because the question was edited mid-flight).
var ErrAudioInputHashMismatch = errors.New("audio input hash mismatch")

// ErrAudioNotPending is returned when a claim is attempted on a question
// whose audio generation is not in the pending state.
var ErrAudioNotPending = errors.New("audio generation is not pending")

// ErrAudioNotGenerating is returned when a complete or fail transition is
// attempted on a question whose audio generation is not in the generating
// state.
var ErrAudioNotGenerating = errors.New("audio generation is not generating")

// ErrAudioConcurrentModification is returned when an audio state transition
// loses an optimistic-lock race against another writer (typically a concurrent
// batch instance or a user edit). Callers should treat it identically to
// ErrAudioNotPending: skip the item and retry on the next batch run.
var ErrAudioConcurrentModification = errors.New("audio generation concurrent modification")
