package question

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	audioGenerationStatusPending    = "pending"
	audioGenerationStatusGenerating = "generating"
	audioGenerationStatusReady      = "ready"
	audioGenerationStatusFailed     = "failed"

	// AudioLangSource is the map key for the audio rendered from the source text.
	AudioLangSource = "source"
	// AudioLangTarget is the map key for the audio rendered from the blank-filled target text.
	AudioLangTarget = "target"

	maxAudioLastErrorLength = 500
)

// AudioGenerationStatus represents the lifecycle state of audio generation
// for a single Question. Initial creation/edit places a question in pending;
// the batch claims it (generating), then transitions to ready or failed.
type AudioGenerationStatus struct {
	value string
}

// AudioGenerationStatusPending returns the pending status.
func AudioGenerationStatusPending() AudioGenerationStatus {
	return AudioGenerationStatus{value: audioGenerationStatusPending}
}

// AudioGenerationStatusGenerating returns the generating status.
func AudioGenerationStatusGenerating() AudioGenerationStatus {
	return AudioGenerationStatus{value: audioGenerationStatusGenerating}
}

// AudioGenerationStatusReady returns the ready status.
func AudioGenerationStatusReady() AudioGenerationStatus {
	return AudioGenerationStatus{value: audioGenerationStatusReady}
}

// AudioGenerationStatusFailed returns the failed status.
func AudioGenerationStatusFailed() AudioGenerationStatus {
	return AudioGenerationStatus{value: audioGenerationStatusFailed}
}

// NewAudioGenerationStatus creates a validated AudioGenerationStatus from a string.
func NewAudioGenerationStatus(value string) (AudioGenerationStatus, error) {
	switch value {
	case audioGenerationStatusPending,
		audioGenerationStatusGenerating,
		audioGenerationStatusReady,
		audioGenerationStatusFailed:
		return AudioGenerationStatus{value: value}, nil
	default:
		return AudioGenerationStatus{}, fmt.Errorf("invalid audio generation status %q: %w", value, domain.ErrInvalidArgument)
	}
}

// Value returns the string representation.
func (s AudioGenerationStatus) Value() string { return s.value }

// IsPending reports whether the status is pending.
func (s AudioGenerationStatus) IsPending() bool { return s.value == audioGenerationStatusPending }

// IsReady reports whether the status is ready (audio is available).
func (s AudioGenerationStatus) IsReady() bool { return s.value == audioGenerationStatusReady }

// AudioRef is a reference to a generated audio file in object storage.
// It carries only enough metadata for the public API to render a playable URL.
type AudioRef struct {
	path        string
	durationSec float64
	sizeBytes   int64
}

// NewAudioRef creates a validated AudioRef.
func NewAudioRef(path string, durationSec float64, sizeBytes int64) (AudioRef, error) {
	if path == "" {
		return AudioRef{}, fmt.Errorf("audio ref path is required: %w", domain.ErrInvalidArgument)
	}
	if durationSec < 0 {
		return AudioRef{}, fmt.Errorf("audio ref duration must not be negative: %w", domain.ErrInvalidArgument)
	}
	if sizeBytes < 0 {
		return AudioRef{}, fmt.Errorf("audio ref size must not be negative: %w", domain.ErrInvalidArgument)
	}
	return AudioRef{path: path, durationSec: durationSec, sizeBytes: sizeBytes}, nil
}

// Path returns the object storage path (relative to the configured bucket).
func (r AudioRef) Path() string { return r.path }

// DurationSec returns the duration of the audio in seconds.
func (r AudioRef) DurationSec() float64 { return r.durationSec }

// SizeBytes returns the file size in bytes.
func (r AudioRef) SizeBytes() int64 { return r.sizeBytes }

// AudioGeneration captures the audio-generation state for a single Question.
// It is updated by the audio batch, not by the question's owner; therefore it
// is intentionally separated from the user-edited Content JSON.
type AudioGeneration struct {
	status         AudioGenerationStatus
	inputHash      string
	refs           map[string]AudioRef
	updatedAt      time.Time
	failedAttempts int
	lastError      string
}

// NewAudioGeneration creates a validated AudioGeneration.
//
// refs may be nil while the audio is still being generated. lastError may be
// empty when the status is not failed.
func NewAudioGeneration(
	status AudioGenerationStatus,
	inputHash string,
	refs map[string]AudioRef,
	updatedAt time.Time,
	failedAttempts int,
	lastError string,
) (*AudioGeneration, error) {
	if status.Value() == "" {
		return nil, fmt.Errorf("audio generation status is required: %w", domain.ErrInvalidArgument)
	}
	if inputHash == "" {
		return nil, fmt.Errorf("audio generation input hash is required: %w", domain.ErrInvalidArgument)
	}
	if failedAttempts < 0 {
		return nil, fmt.Errorf("audio generation failed attempts must not be negative: %w", domain.ErrInvalidArgument)
	}
	if len(lastError) > maxAudioLastErrorLength {
		return nil, fmt.Errorf("audio generation last error must not exceed %d characters: %w", maxAudioLastErrorLength, domain.ErrInvalidArgument)
	}
	return &AudioGeneration{
		status:         status,
		inputHash:      inputHash,
		refs:           copyAudioRefs(refs),
		updatedAt:      updatedAt,
		failedAttempts: failedAttempts,
		lastError:      lastError,
	}, nil
}

// Status returns the current generation status.
func (a *AudioGeneration) Status() AudioGenerationStatus { return a.status }

// InputHash returns the hash of the input text that produced (or will produce) the audio.
func (a *AudioGeneration) InputHash() string { return a.inputHash }

// Refs returns a defensive copy of the audio file refs keyed by language slot
// (AudioLangSource / AudioLangTarget).
func (a *AudioGeneration) Refs() map[string]AudioRef { return copyAudioRefs(a.refs) }

// UpdatedAt returns the last status update timestamp.
func (a *AudioGeneration) UpdatedAt() time.Time { return a.updatedAt }

// FailedAttempts returns the consecutive failure count.
func (a *AudioGeneration) FailedAttempts() int { return a.failedAttempts }

// LastError returns the last failure message (empty when status != failed).
func (a *AudioGeneration) LastError() string { return a.lastError }

func copyAudioRefs(refs map[string]AudioRef) map[string]AudioRef {
	if refs == nil {
		return nil
	}
	cp := make(map[string]AudioRef, len(refs))
	maps.Copy(cp, refs)
	return cp
}

// ComputeWordFillAudioInputHash returns a deterministic hash of the inputs that
// drive audio generation for a word_fill question.
//
// The blank placeholders ({{...}}) in the target text are replaced with the
// inner word so the hash matches what the TTS engine actually receives. Each
// component is separated by a NUL byte to guarantee field boundaries.
func ComputeWordFillAudioInputHash(c WordFillContent) string {
	targetFilled := FillWordFillBlanks(c.Target.Text)
	h := sha256.New()
	h.Write([]byte(c.Source.Text))
	h.Write([]byte{0})
	h.Write([]byte(c.Source.Lang))
	h.Write([]byte{0})
	h.Write([]byte(targetFilled))
	h.Write([]byte{0})
	h.Write([]byte(c.Target.Lang))
	return hex.EncodeToString(h.Sum(nil))
}

// FillWordFillBlanks replaces every {{word}} placeholder in the given text with
// its trimmed inner content.
func FillWordFillBlanks(text string) string {
	return blankPattern.ReplaceAllStringFunc(text, func(match string) string {
		inner := match[2 : len(match)-2]
		return strings.TrimSpace(inner)
	})
}

// truncateRunes returns at most maxRunes runes of s. Slicing on bytes would
// split multi-byte UTF-8 characters mid-codepoint and corrupt the stored value,
// so truncation is done on the rune sequence.
func truncateRunes(s string, maxRunes int) string {
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes])
}
