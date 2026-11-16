package question_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainquestion "github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

func Test_AudioOutputFromRefs_shouldReturnNil_whenNoSlotHasRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		refs map[string]domainquestion.AudioRef
	}{
		{name: "nil", refs: nil},
		{name: "empty", refs: map[string]domainquestion.AudioRef{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			out := question.AudioOutputFromRefs(tt.refs)

			// then
			assert.Nil(t, out)
		})
	}
}

func Test_AudioOutputFromRefs_shouldReturnSourceOnly_whenOnlySourceHasRef(t *testing.T) {
	t.Parallel()

	// given
	refs := map[string]domainquestion.AudioRef{
		domainquestion.AudioLangSource: fixtureAudioRef(t, "audio/q1/source.opus", 1.5),
	}

	// when
	out := question.AudioOutputFromRefs(refs)

	// then
	assert.Equal(t, &question.AudioOutput{
		Source: &question.AudioRefOutput{Path: "audio/q1/source.opus", DurationSec: 1.5},
		Target: nil,
	}, out)
}

func Test_AudioOutputFromRefs_shouldReturnTargetOnly_whenOnlyTargetHasRef(t *testing.T) {
	t.Parallel()

	// given
	refs := map[string]domainquestion.AudioRef{
		domainquestion.AudioLangTarget: fixtureAudioRef(t, "audio/q1/target.opus", 2.5),
	}

	// when
	out := question.AudioOutputFromRefs(refs)

	// then
	assert.Equal(t, &question.AudioOutput{
		Source: nil,
		Target: &question.AudioRefOutput{Path: "audio/q1/target.opus", DurationSec: 2.5},
	}, out)
}

func fixtureAudioRef(t *testing.T, path string, durationSec float64) domainquestion.AudioRef {
	t.Helper()
	ref, err := domainquestion.NewAudioRef(path, durationSec, 100)
	require.NoError(t, err)
	return ref
}
