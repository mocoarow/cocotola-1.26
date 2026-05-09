package question_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

const (
	fixtureQuestionID = "q-1"
	fixtureWorkbookID = "wb-1"
)

type questionArgs struct {
	id         string
	workbookID string
	qt         question.Type
	content    string
	tags       []string
	orderIndex int
	createdAt  time.Time
	updatedAt  time.Time
}

func validQuestionArgs() questionArgs {
	now := time.Now()
	return questionArgs{
		id:         fixtureQuestionID,
		workbookID: fixtureWorkbookID,
		qt:         question.TypeWordFill(),
		content:    `{"source":{"text":"りんご","lang":"ja"},"target":{"text":"{{apple}}","lang":"en"}}`,
		tags:       nil,
		orderIndex: 0,
		createdAt:  now,
		updatedAt:  now,
	}
}

func Test_NewQuestion_shouldReturnQuestion_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()

	// when
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, a.id, q.ID())
	assert.Equal(t, a.workbookID, q.WorkbookID())
	assert.Equal(t, "word_fill", q.QuestionType().Value())
	assert.Equal(t, a.content, q.Content())
	assert.Equal(t, a.orderIndex, q.OrderIndex())
	assert.Equal(t, 0, q.Version())
	assert.Equal(t, a.createdAt, q.CreatedAt())
	assert.Equal(t, a.updatedAt, q.UpdatedAt())
}

func Test_NewQuestion_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()

	// when
	_, err := question.NewQuestion("", a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenWorkbookIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()

	// when
	_, err := question.NewQuestion(a.id, "", a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenQuestionTypeIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, question.Type{}, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenContentIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, a.qt, "", a.tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenContentExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	longContent := strings.Repeat("a", 10001)

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, a.qt, longContent, a.tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenOrderIndexIsNegative(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, -1, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnQuestion_whenTagsAreValid(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	tags := []string{"level:beginner", "topic:grammar"}

	// when
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, tags, q.Tags())
}

func Test_NewQuestion_shouldReturnError_whenTagsExceedMax(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	tags := make([]string, 21)
	for i := range tags {
		tags[i] = "key:value"
	}

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenTagFormatIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tag  string
	}{
		{name: "no_colon", tag: "invalidtag"},
		{name: "empty_key", tag: ":value"},
		{name: "empty_value", tag: "key:"},
		{name: "spaces", tag: "key: value"},
		{name: "special_chars", tag: "key:val@ue"},
		{name: "multiple_colons", tag: "seed:wb-v1:q1"},
		{name: "trailing_colon_segment", tag: "key:value:extra"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			a := validQuestionArgs()

			// when
			_, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, []string{tt.tag}, a.orderIndex, a.createdAt, a.updatedAt)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewQuestion_shouldReturnError_whenTagIsDuplicated(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	tags := []string{"level:beginner", "level:beginner"}

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, tags, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenTagExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	longTag := strings.Repeat("a", 50) + ":" + strings.Repeat("b", 50)

	// when
	_, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, []string{longTag}, a.orderIndex, a.createdAt, a.updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Tags_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	tags := []string{"level:beginner"}
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)

	// when
	returned := q.Tags()
	returned[0] = "mutated:value"

	// then
	assert.Equal(t, "level:beginner", q.Tags()[0])
}

func Test_ReconstructQuestion_shouldReturnQuestion_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now()
	id := "q-1"
	workbookID := "wb-1"
	qt := question.TypeWordFill()
	content := `{"source":{"text":"りんご","lang":"ja"},"target":{"text":"{{apple}}","lang":"en"}}`
	tags := []string{"level:beginner"}
	orderIndex := 3
	version := 7

	// when
	q := question.ReconstructQuestion(id, workbookID, qt, content, tags, orderIndex, version, now, now)

	// then
	assert.Equal(t, id, q.ID())
	assert.Equal(t, workbookID, q.WorkbookID())
	assert.Equal(t, content, q.Content())
	assert.Equal(t, tags, q.Tags())
	assert.Equal(t, orderIndex, q.OrderIndex())
	assert.Equal(t, version, q.Version())
}

func Test_SetVersion_shouldUpdateVersion(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)

	// when
	q.SetVersion(5)

	// then
	assert.Equal(t, 5, q.Version())
}

func Test_Edit_shouldUpdateFields_whenInputIsValid(t *testing.T) {
	t.Parallel()

	// given
	originalContent := `{"source":{"text":"original","lang":"ja"},"target":{"text":"{{a}}","lang":"en"}}`
	originalUpdatedAt := time.Now().Add(-time.Hour)
	q := question.ReconstructQuestion("q-1", fixtureWorkbookID, question.TypeWordFill(), originalContent, nil, 0, 0, originalUpdatedAt, originalUpdatedAt)

	newContent := `{"source":{"text":"updated","lang":"ja"},"target":{"text":"{{b}}","lang":"en"}}`
	newTags := []string{"level:advanced"}
	newUpdatedAt := originalUpdatedAt.Add(time.Hour)

	// when
	err := q.Edit(newContent, newTags, 5, newUpdatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, newContent, q.Content())
	assert.Equal(t, newTags, q.Tags())
	assert.Equal(t, 5, q.OrderIndex())
	assert.Equal(t, newUpdatedAt, q.UpdatedAt())
}

func Test_Edit_shouldNotMutateState_whenContentIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, []string{"level:beginner"}, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	originalContent := q.Content()
	originalTags := q.Tags()
	originalOrderIndex := q.OrderIndex()
	originalUpdatedAt := q.UpdatedAt()

	// when
	err = q.Edit("", []string{"level:advanced"}, 5, originalUpdatedAt.Add(time.Hour))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
	assert.Equal(t, originalContent, q.Content())
	assert.Equal(t, originalTags, q.Tags())
	assert.Equal(t, originalOrderIndex, q.OrderIndex())
	assert.Equal(t, originalUpdatedAt, q.UpdatedAt())
}

func Test_Edit_shouldReturnError_whenOrderIndexIsNegative(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, nil, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)

	// when
	err = q.Edit(a.content, nil, -1, a.updatedAt.Add(time.Hour))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Question_AudioGeneration_shouldReturnNil_whenUnset(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)

	// then
	assert.Nil(t, q.AudioGeneration())
}

func Test_Question_MarkAudioPending_shouldSetPending_whenInitiallyNil(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	now := time.Now()

	// when
	q.MarkAudioPending("hash-v1", now)

	// then
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "pending", ag.Status().Value())
	assert.Equal(t, "hash-v1", ag.InputHash())
	assert.Equal(t, now, ag.UpdatedAt())
	assert.Equal(t, 0, ag.FailedAttempts())
	assert.Empty(t, ag.LastError())
	assert.Nil(t, ag.Refs())
}

func Test_Question_MarkAudioPending_shouldBeNoOp_whenInputHashUnchanged(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	first := time.Now()
	q.MarkAudioPending("hash-v1", first)

	// when: same inputHash, later timestamp
	q.MarkAudioPending("hash-v1", first.Add(time.Hour))

	// then: original state preserved
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, first, ag.UpdatedAt())
}

func Test_Question_MarkAudioPending_shouldResetState_whenInputHashChanges(t *testing.T) {
	t.Parallel()

	// given: a question already in ready state
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	ref, err := question.NewAudioRef("audio/questions/x/source.opus", 1.0, 100)
	require.NoError(t, err)
	prev, err := question.NewAudioGeneration(
		question.AudioGenerationStatusReady(),
		"hash-v0",
		map[string]question.AudioRef{question.AudioLangSource: ref},
		time.Now(),
		2,
		"",
	)
	require.NoError(t, err)
	q.SetAudioGeneration(prev)

	// when
	q.MarkAudioPending("hash-v1", time.Now())

	// then: refs and counters are cleared, new hash applied
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "pending", ag.Status().Value())
	assert.Equal(t, "hash-v1", ag.InputHash())
	assert.Equal(t, 0, ag.FailedAttempts())
	assert.Nil(t, ag.Refs())
}

// fixtureAudioGeneration returns a *AudioGeneration with the supplied status
// for use in audio state-transition tests.
func fixtureAudioGeneration(t *testing.T, status question.AudioGenerationStatus, inputHash string, updatedAt time.Time, failedAttempts int) *question.AudioGeneration {
	t.Helper()
	ag, err := question.NewAudioGeneration(status, inputHash, nil, updatedAt, failedAttempts, "")
	require.NoError(t, err)
	return ag
}

func Test_Question_ClaimAudio_shouldTransitionToGenerating_whenPending(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	queuedAt := time.Now()
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusPending(), "hash-v1", queuedAt, 0))
	claimedAt := queuedAt.Add(time.Minute)

	// when
	err = q.ClaimAudio("hash-v1", claimedAt)

	// then
	require.NoError(t, err)
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "generating", ag.Status().Value())
	assert.Equal(t, "hash-v1", ag.InputHash())
	assert.Equal(t, claimedAt, ag.UpdatedAt())
	assert.Empty(t, ag.LastError())
}

func Test_Question_ClaimAudio_shouldReturnError_whenAudioGenerationIsNil(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)

	// when
	err = q.ClaimAudio("hash-v1", time.Now())

	// then
	require.ErrorIs(t, err, domain.ErrAudioNotPending)
}

func Test_Question_ClaimAudio_shouldReturnError_whenStatusIsNotPending(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status question.AudioGenerationStatus
	}{
		{name: "generating", status: question.AudioGenerationStatusGenerating()},
		{name: "ready", status: question.AudioGenerationStatusReady()},
		{name: "failed", status: question.AudioGenerationStatusFailed()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			a := validQuestionArgs()
			q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
			require.NoError(t, err)
			q.SetAudioGeneration(fixtureAudioGeneration(t, tt.status, "hash-v1", time.Now(), 0))

			// when
			err = q.ClaimAudio("hash-v1", time.Now())

			// then
			require.ErrorIs(t, err, domain.ErrAudioNotPending)
		})
	}
}

func Test_Question_ClaimAudio_shouldReturnError_whenInputHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusPending(), "hash-v1", time.Now(), 0))

	// when
	err = q.ClaimAudio("hash-v2", time.Now())

	// then
	require.ErrorIs(t, err, domain.ErrAudioInputHashMismatch)
}

func Test_Question_CompleteAudio_shouldTransitionToReady_whenGenerating(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", time.Now(), 1))
	srcRef, err := question.NewAudioRef("audio/questions/q1/source.opus", 2.5, 1234)
	require.NoError(t, err)
	tgtRef, err := question.NewAudioRef("audio/questions/q1/target.opus", 3.5, 5678)
	require.NoError(t, err)
	completedAt := time.Now().Add(time.Minute)

	// when
	err = q.CompleteAudio("hash-v1", map[string]question.AudioRef{
		question.AudioLangSource: srcRef,
		question.AudioLangTarget: tgtRef,
	}, completedAt)

	// then
	require.NoError(t, err)
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "ready", ag.Status().Value())
	assert.Equal(t, completedAt, ag.UpdatedAt())
	assert.Equal(t, 0, ag.FailedAttempts())
	assert.Empty(t, ag.LastError())
	refs := ag.Refs()
	assert.Equal(t, srcRef.Path(), refs[question.AudioLangSource].Path())
	assert.Equal(t, tgtRef.Path(), refs[question.AudioLangTarget].Path())
}

func Test_Question_CompleteAudio_shouldReturnError_whenStatusIsNotGenerating(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status question.AudioGenerationStatus
	}{
		{name: "pending", status: question.AudioGenerationStatusPending()},
		{name: "ready", status: question.AudioGenerationStatusReady()},
		{name: "failed", status: question.AudioGenerationStatusFailed()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			a := validQuestionArgs()
			q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
			require.NoError(t, err)
			q.SetAudioGeneration(fixtureAudioGeneration(t, tt.status, "hash-v1", time.Now(), 0))

			// when
			err = q.CompleteAudio("hash-v1", nil, time.Now())

			// then
			require.ErrorIs(t, err, domain.ErrAudioNotGenerating)
		})
	}
}

func Test_Question_CompleteAudio_shouldReturnError_whenInputHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", time.Now(), 0))

	// when
	err = q.CompleteAudio("hash-v2", nil, time.Now())

	// then
	require.ErrorIs(t, err, domain.ErrAudioInputHashMismatch)
}

func Test_Question_FailAudio_shouldTransitionToFailed_whenGenerating(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", time.Now(), 2))
	failedAt := time.Now().Add(time.Minute)

	// when
	err = q.FailAudio("hash-v1", "tts: invalid voice", failedAt)

	// then
	require.NoError(t, err)
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "failed", ag.Status().Value())
	assert.Equal(t, "hash-v1", ag.InputHash())
	assert.Equal(t, failedAt, ag.UpdatedAt())
	assert.Equal(t, 3, ag.FailedAttempts())
	assert.Equal(t, "tts: invalid voice", ag.LastError())
}

func Test_Question_FailAudio_shouldTruncateReason_whenReasonExceedsMaxRunes(t *testing.T) {
	t.Parallel()

	// given: 600 multi-byte runes (Japanese hiragana 'あ' is 3 bytes each).
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", time.Now(), 0))
	reason := strings.Repeat("あ", 600)

	// when
	err = q.FailAudio("hash-v1", reason, time.Now())

	// then
	require.NoError(t, err)
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	last := ag.LastError()
	assert.Len(t, []rune(last), 500)
	assert.Equal(t, strings.Repeat("あ", 500), last)
}

func Test_Question_FailAudio_shouldReturnError_whenStatusIsNotGenerating(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status question.AudioGenerationStatus
	}{
		{name: "pending", status: question.AudioGenerationStatusPending()},
		{name: "ready", status: question.AudioGenerationStatusReady()},
		{name: "failed", status: question.AudioGenerationStatusFailed()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			a := validQuestionArgs()
			q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
			require.NoError(t, err)
			q.SetAudioGeneration(fixtureAudioGeneration(t, tt.status, "hash-v1", time.Now(), 0))

			// when
			err = q.FailAudio("hash-v1", "boom", time.Now())

			// then
			require.ErrorIs(t, err, domain.ErrAudioNotGenerating)
		})
	}
}

func Test_Question_FailAudio_shouldReturnError_whenInputHashDoesNotMatch(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", time.Now(), 0))

	// when
	err = q.FailAudio("hash-v2", "boom", time.Now())

	// then
	require.ErrorIs(t, err, domain.ErrAudioInputHashMismatch)
}

func Test_Question_ReclaimStaleAudio_shouldReturnFalse_whenAudioGenerationIsNil(t *testing.T) {
	t.Parallel()

	// given
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)

	// when
	reclaimed := q.ReclaimStaleAudio(time.Now(), time.Minute)

	// then
	assert.False(t, reclaimed)
	assert.Nil(t, q.AudioGeneration())
}

func Test_Question_ReclaimStaleAudio_shouldReturnFalse_whenStatusIsNotGenerating(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status question.AudioGenerationStatus
	}{
		{name: "pending", status: question.AudioGenerationStatusPending()},
		{name: "ready", status: question.AudioGenerationStatusReady()},
		{name: "failed", status: question.AudioGenerationStatusFailed()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given: stale by clock but wrong status
			a := validQuestionArgs()
			q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
			require.NoError(t, err)
			old := time.Now().Add(-time.Hour)
			q.SetAudioGeneration(fixtureAudioGeneration(t, tt.status, "hash-v1", old, 0))

			// when
			reclaimed := q.ReclaimStaleAudio(time.Now(), time.Minute)

			// then
			assert.False(t, reclaimed)
			assert.Equal(t, tt.status.Value(), q.AudioGeneration().Status().Value())
		})
	}
}

func Test_Question_ReclaimStaleAudio_shouldReturnFalse_whenNotStale(t *testing.T) {
	t.Parallel()

	// given: generating but updated only 10s ago
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	now := time.Now()
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", now.Add(-10*time.Second), 0))

	// when
	reclaimed := q.ReclaimStaleAudio(now, time.Minute)

	// then
	assert.False(t, reclaimed)
	assert.Equal(t, "generating", q.AudioGeneration().Status().Value())
}

func Test_Question_ReclaimStaleAudio_shouldReturnTrueAndResetToPending_whenStale(t *testing.T) {
	t.Parallel()

	// given: generating updated 10 minutes ago, staleAfter 1 minute
	a := validQuestionArgs()
	q, err := question.NewQuestion(a.id, a.workbookID, a.qt, a.content, a.tags, a.orderIndex, a.createdAt, a.updatedAt)
	require.NoError(t, err)
	now := time.Now()
	q.SetAudioGeneration(fixtureAudioGeneration(t, question.AudioGenerationStatusGenerating(), "hash-v1", now.Add(-10*time.Minute), 4))

	// when
	reclaimed := q.ReclaimStaleAudio(now, time.Minute)

	// then
	assert.True(t, reclaimed)
	ag := q.AudioGeneration()
	require.NotNil(t, ag)
	assert.Equal(t, "pending", ag.Status().Value())
	assert.Equal(t, "hash-v1", ag.InputHash())
	assert.Equal(t, now, ag.UpdatedAt())
	assert.Equal(t, 4, ag.FailedAttempts(), "failedAttempts must be preserved across reclaim")
}
