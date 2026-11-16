package handler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/api"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/controller/handler"
	questionservice "github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

func Test_AudioURLBuilder_QuestionAudio_shouldReturnEmptyRefs_whenAudioIsNil(t *testing.T) {
	t.Parallel()

	// given
	b := handler.NewAudioURLBuilder("https://cdn.example.com/audio")

	// when
	out := b.QuestionAudio(nil)

	// then
	assert.Equal(t, api.QuestionAudio{
		Source: api.QuestionAudioRef{Url: "", DurationSec: 0},
		Target: api.QuestionAudioRef{Url: "", DurationSec: 0},
	}, out)
}

func Test_AudioURLBuilder_QuestionAudio_shouldReturnURLsAndDurations_whenBothSlotsAreSet(t *testing.T) {
	t.Parallel()

	// given
	b := handler.NewAudioURLBuilder("https://cdn.example.com/audio")
	audio := &questionservice.AudioOutput{
		Source: &questionservice.AudioRefOutput{Path: "q1/source.opus", DurationSec: 1.5},
		Target: &questionservice.AudioRefOutput{Path: "q1/target.opus", DurationSec: 2.5},
	}

	// when
	out := b.QuestionAudio(audio)

	// then
	assert.Equal(t, api.QuestionAudio{
		Source: api.QuestionAudioRef{Url: "https://cdn.example.com/audio/q1/source.opus", DurationSec: 1.5},
		Target: api.QuestionAudioRef{Url: "https://cdn.example.com/audio/q1/target.opus", DurationSec: 2.5},
	}, out)
}

func Test_AudioURLBuilder_QuestionAudio_shouldReturnEmptyRef_whenSlotIsNil(t *testing.T) {
	t.Parallel()

	// given
	b := handler.NewAudioURLBuilder("https://cdn.example.com/audio")
	audio := &questionservice.AudioOutput{
		Source: nil,
		Target: &questionservice.AudioRefOutput{Path: "q1/target.opus", DurationSec: 2.5},
	}

	// when
	out := b.QuestionAudio(audio)

	// then
	assert.Equal(t, api.QuestionAudioRef{Url: "", DurationSec: 0}, out.Source)
}

func Test_AudioURLBuilder_QuestionAudio_shouldJoinWithSingleSlash_whenSlashesSurroundTheJoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		path    string
	}{
		{name: "no slashes", baseURL: "https://cdn.example.com/audio", path: "q1/source.opus"},
		{name: "trailing slash on base", baseURL: "https://cdn.example.com/audio/", path: "q1/source.opus"},
		{name: "leading slash on path", baseURL: "https://cdn.example.com/audio", path: "/q1/source.opus"},
		{name: "repeated slashes on both", baseURL: "https://cdn.example.com/audio//", path: "//q1/source.opus"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			b := handler.NewAudioURLBuilder(tt.baseURL)
			audio := &questionservice.AudioOutput{
				Source: &questionservice.AudioRefOutput{Path: tt.path, DurationSec: 1.5},
				Target: nil,
			}

			// when
			out := b.QuestionAudio(audio)

			// then
			assert.Equal(t, "https://cdn.example.com/audio/q1/source.opus", out.Source.Url)
		})
	}
}

func Test_AudioURLBuilder_QuestionAudio_shouldReturnEmptyURL_whenPathIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	b := handler.NewAudioURLBuilder("https://cdn.example.com/audio")
	audio := &questionservice.AudioOutput{
		Source: &questionservice.AudioRefOutput{Path: "", DurationSec: 1.5},
		Target: nil,
	}

	// when
	out := b.QuestionAudio(audio)

	// then
	assert.Empty(t, out.Source.Url)
}
