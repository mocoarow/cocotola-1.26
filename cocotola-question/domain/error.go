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

// ErrInvalidSpaceType is returned when cocotola-auth returns an unknown SpaceType.
var ErrInvalidSpaceType = errors.New("invalid space type")

// ErrAudioInputHashMismatch is returned when the question was edited after the audio job started.
var ErrAudioInputHashMismatch = errors.New("audio input hash mismatch")

// ErrAudioNotPending is returned when claiming audio that is not pending.
var ErrAudioNotPending = errors.New("audio generation is not pending")

// ErrAudioNotGenerating is returned when completing or failing audio that is not generating.
var ErrAudioNotGenerating = errors.New("audio generation is not generating")

// ErrAudioConcurrentModification is an optimistic-lock loss; handle like ErrAudioNotPending (skip, retry next run).
var ErrAudioConcurrentModification = errors.New("audio generation concurrent modification")
