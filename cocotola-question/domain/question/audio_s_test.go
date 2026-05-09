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

func Test_NewAudioGenerationStatus_shouldReturnStatus_whenValueIsKnown(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{name: "pending", value: "pending"},
		{name: "generating", value: "generating"},
		{name: "ready", value: "ready"},
		{name: "failed", value: "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			s, err := question.NewAudioGenerationStatus(tt.value)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.value, s.Value())
		})
	}
}

func Test_NewAudioGenerationStatus_shouldReturnError_whenValueIsUnknown(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioGenerationStatus("done")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_AudioGenerationStatus_IsPending_shouldReturnTrue_whenPending(t *testing.T) {
	t.Parallel()

	// given
	s := question.AudioGenerationStatusPending()

	// then
	assert.True(t, s.IsPending())
	assert.False(t, s.IsReady())
}

func Test_AudioGenerationStatus_IsReady_shouldReturnTrue_whenReady(t *testing.T) {
	t.Parallel()

	// given
	s := question.AudioGenerationStatusReady()

	// then
	assert.True(t, s.IsReady())
	assert.False(t, s.IsPending())
}

func Test_NewAudioRef_shouldReturnRef_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// when
	r, err := question.NewAudioRef("audio/questions/q1/source.opus", 5.2, 24000)

	// then
	require.NoError(t, err)
	assert.Equal(t, "audio/questions/q1/source.opus", r.Path())
	assert.InEpsilon(t, 5.2, r.DurationSec(), 1e-9)
	assert.Equal(t, int64(24000), r.SizeBytes())
}

func Test_NewAudioRef_shouldReturnError_whenPathIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioRef("", 1.0, 100)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewAudioRef_shouldReturnError_whenDurationIsNegative(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioRef("path", -0.1, 100)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewAudioRef_shouldReturnError_whenSizeIsNegative(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioRef("path", 1.0, -1)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewAudioGeneration_shouldReturnAggregate_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	now := time.Now()
	ref, err := question.NewAudioRef("audio/questions/q1/source.opus", 4.3, 16384)
	require.NoError(t, err)

	// when
	a, err := question.NewAudioGeneration(
		question.AudioGenerationStatusReady(),
		"abc123",
		map[string]question.AudioRef{question.AudioLangSource: ref},
		now,
		0,
		"",
	)

	// then
	require.NoError(t, err)
	assert.Equal(t, "ready", a.Status().Value())
	assert.Equal(t, "abc123", a.InputHash())
	assert.Equal(t, now, a.UpdatedAt())
	assert.Equal(t, 0, a.FailedAttempts())
	assert.Empty(t, a.LastError())
	refs := a.Refs()
	require.Contains(t, refs, question.AudioLangSource)
	assert.Equal(t, ref.Path(), refs[question.AudioLangSource].Path())
}

func Test_NewAudioGeneration_shouldReturnError_whenStatusIsZero(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioGeneration(
		question.AudioGenerationStatus{},
		"hash",
		nil,
		time.Now(),
		0,
		"",
	)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewAudioGeneration_shouldReturnError_whenInputHashIsEmpty(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioGeneration(
		question.AudioGenerationStatusPending(),
		"",
		nil,
		time.Now(),
		0,
		"",
	)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewAudioGeneration_shouldReturnError_whenFailedAttemptsIsNegative(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.NewAudioGeneration(
		question.AudioGenerationStatusFailed(),
		"hash",
		nil,
		time.Now(),
		-1,
		"boom",
	)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewAudioGeneration_shouldReturnError_whenLastErrorTooLong(t *testing.T) {
	t.Parallel()

	// given
	long := strings.Repeat("x", 501)

	// when
	_, err := question.NewAudioGeneration(
		question.AudioGenerationStatusFailed(),
		"hash",
		nil,
		time.Now(),
		1,
		long,
	)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_AudioGeneration_Refs_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	ref, err := question.NewAudioRef("audio/questions/q1/source.opus", 1.0, 100)
	require.NoError(t, err)
	a, err := question.NewAudioGeneration(
		question.AudioGenerationStatusReady(),
		"hash",
		map[string]question.AudioRef{question.AudioLangSource: ref},
		time.Now(),
		0,
		"",
	)
	require.NoError(t, err)

	// when: caller mutates the returned map
	out := a.Refs()
	delete(out, question.AudioLangSource)

	// then: aggregate is unaffected
	assert.Contains(t, a.Refs(), question.AudioLangSource)
}

func Test_FillWordFillBlanks_shouldReplaceBracesWithInnerWord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "single_blank", input: "I {{eat}} apples", want: "I eat apples"},
		{name: "multiple_blanks", input: "I {{travel}} to {{Tokyo}}", want: "I travel to Tokyo"},
		{name: "blank_with_spaces", input: "open the {{ window }}", want: "open the window"},
		{name: "no_blanks", input: "plain text", want: "plain text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			got := question.FillWordFillBlanks(tt.input)

			// then
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_ComputeWordFillAudioInputHash_shouldBeDeterministic(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんごを食べる", Lang: "ja"},
		Target: question.TextWithLang{Text: "eat an {{apple}}", Lang: "en"},
	}

	// when
	got1 := question.ComputeWordFillAudioInputHash(c)
	got2 := question.ComputeWordFillAudioInputHash(c)

	// then
	assert.Equal(t, got1, got2)
	assert.Len(t, got1, 64) // sha256 hex
}

func Test_ComputeWordFillAudioInputHash_shouldChange_whenAnyFieldChanges(t *testing.T) {
	t.Parallel()

	base := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんごを食べる", Lang: "ja"},
		Target: question.TextWithLang{Text: "eat an {{apple}}", Lang: "en"},
	}
	baseHash := question.ComputeWordFillAudioInputHash(base)

	tests := []struct {
		name string
		mut  func(*question.WordFillContent)
	}{
		{name: "source_text", mut: func(c *question.WordFillContent) { c.Source.Text = "ばななを食べる" }},
		{name: "source_lang", mut: func(c *question.WordFillContent) { c.Source.Lang = "en" }},
		{name: "target_text", mut: func(c *question.WordFillContent) { c.Target.Text = "eat a {{banana}}" }},
		{name: "target_lang", mut: func(c *question.WordFillContent) { c.Target.Lang = "fr" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			c := base
			tt.mut(&c)

			// when
			got := question.ComputeWordFillAudioInputHash(c)

			// then
			assert.NotEqual(t, baseHash, got)
		})
	}
}

func Test_ComputeWordFillAudioInputHash_shouldDeriveSameHash_whenBlanksAreEquivalent(t *testing.T) {
	t.Parallel()

	// given
	withBlank := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんごを食べる", Lang: "ja"},
		Target: question.TextWithLang{Text: "eat an {{apple}}", Lang: "en"},
	}
	noBlank := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんごを食べる", Lang: "ja"},
		Target: question.TextWithLang{Text: "eat an apple", Lang: "en"},
	}

	// when
	got1 := question.ComputeWordFillAudioInputHash(withBlank)
	got2 := question.ComputeWordFillAudioInputHash(noBlank)

	// then: filling the blank produces the same surface text TTS sees
	assert.Equal(t, got1, got2)
}
