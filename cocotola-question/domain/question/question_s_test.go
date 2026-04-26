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

func validQuestionArgs() (string, question.Type, string, []string, int, time.Time, time.Time) {
	now := time.Now()
	content := `{"source":{"text":"りんご","lang":"ja"},"target":{"text":"{{apple}}","lang":"en"}}`
	return "q-1", question.TypeWordFill(), content, nil, 0, now, now
}

func Test_NewQuestion_shouldReturnQuestion_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, tags, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	q, err := question.NewQuestion(id, qt, content, tags, orderIndex, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, q.ID())
	assert.Equal(t, "word_fill", q.QuestionType().Value())
	assert.Equal(t, content, q.Content())
	assert.Equal(t, orderIndex, q.OrderIndex())
	assert.Equal(t, createdAt, q.CreatedAt())
	assert.Equal(t, updatedAt, q.UpdatedAt())
}

func Test_NewQuestion_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, qt, content, tags, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion("", qt, content, tags, orderIndex, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenQuestionTypeIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, _, content, tags, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion(id, question.Type{}, content, tags, orderIndex, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenContentIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, qt, _, tags, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion(id, qt, "", tags, orderIndex, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenContentExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, qt, _, tags, orderIndex, createdAt, updatedAt := validQuestionArgs()
	longContent := strings.Repeat("a", 10001)

	// when
	_, err := question.NewQuestion(id, qt, longContent, tags, orderIndex, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenOrderIndexIsNegative(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, tags, _, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion(id, qt, content, tags, -1, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnQuestion_whenTagsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, _, orderIndex, createdAt, updatedAt := validQuestionArgs()
	tags := []string{"level:beginner", "topic:grammar"}

	// when
	q, err := question.NewQuestion(id, qt, content, tags, orderIndex, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, tags, q.Tags())
}

func Test_NewQuestion_shouldReturnError_whenTagsExceedMax(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, _, orderIndex, createdAt, updatedAt := validQuestionArgs()
	tags := make([]string, 21)
	for i := range tags {
		tags[i] = "key:value"
	}

	// when
	_, err := question.NewQuestion(id, qt, content, tags, orderIndex, createdAt, updatedAt)

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
			id, qt, content, _, orderIndex, createdAt, updatedAt := validQuestionArgs()

			// when
			_, err := question.NewQuestion(id, qt, content, []string{tt.tag}, orderIndex, createdAt, updatedAt)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewQuestion_shouldReturnError_whenTagIsDuplicated(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, _, orderIndex, createdAt, updatedAt := validQuestionArgs()
	tags := []string{"level:beginner", "level:beginner"}

	// when
	_, err := question.NewQuestion(id, qt, content, tags, orderIndex, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_NewQuestion_shouldReturnError_whenTagExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, _, orderIndex, createdAt, updatedAt := validQuestionArgs()
	longTag := strings.Repeat("a", 50) + ":" + strings.Repeat("b", 50)

	// when
	_, err := question.NewQuestion(id, qt, content, []string{longTag}, orderIndex, createdAt, updatedAt)

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Tags_shouldReturnDefensiveCopy(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, _, orderIndex, createdAt, updatedAt := validQuestionArgs()
	tags := []string{"level:beginner"}
	q, err := question.NewQuestion(id, qt, content, tags, orderIndex, createdAt, updatedAt)
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
	id, qt, content, tags, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	q := question.ReconstructQuestion(id, qt, content, tags, orderIndex, createdAt, updatedAt)

	// then
	assert.Equal(t, id, q.ID())
	assert.Equal(t, content, q.Content())
}
