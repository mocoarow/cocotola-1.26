// Package domain holds the small set of value types used internally by
// cocotola-audio-generator. They mirror the cocotola-question audio domain
// without sharing types directly to keep this service decoupled.
package domain

import "errors"

// PendingItem is one queue entry returned by cocotola-question's pending list
// endpoint. The batch will run TTS on (SourceText, SourceLang) and
// (TargetText, TargetLang).
type PendingItem struct {
	WorkbookID  string
	QuestionID  string
	SourceText  string
	SourceLang  string
	TargetText  string
	TargetLang  string
	InputHash   string
	FailedTries int
}

// AudioRef describes a generated audio file ready for upload.
type AudioRef struct {
	Path        string
	DurationSec float64
	SizeBytes   int64
}

// SlotSource is the map key for the audio rendered from the source text.
const SlotSource = "source"

// SlotTarget is the map key for the audio rendered from the blank-filled target text.
const SlotTarget = "target"

// ErrNoPendingItems indicates a batch tick with no work to do.
var ErrNoPendingItems = errors.New("no pending audio items")

// ErrClaimRace indicates another batch instance won the race to claim the
// item. The caller should skip this item and continue with the next one.
var ErrClaimRace = errors.New("audio claim lost race")
