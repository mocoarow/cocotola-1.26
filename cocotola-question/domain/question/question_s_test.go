package question_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func validQuestionArgs() (string, question.Type, string, int, time.Time, time.Time) {
	now := time.Now()
	return "q-1", question.TypeDefault(), `{"text":"hello"}`, 0, now, now
}

func Test_NewQuestion_shouldReturnQuestion_whenAllFieldsAreValid(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	q, err := question.NewQuestion(id, qt, content, orderIndex, createdAt, updatedAt)

	// then
	require.NoError(t, err)
	assert.Equal(t, id, q.ID())
	assert.Equal(t, "default", q.QuestionType().Value())
	assert.Equal(t, content, q.Content())
	assert.Equal(t, orderIndex, q.OrderIndex())
	assert.Equal(t, createdAt, q.CreatedAt())
	assert.Equal(t, updatedAt, q.UpdatedAt())
}

func Test_NewQuestion_shouldReturnError_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	_, qt, content, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion("", qt, content, orderIndex, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewQuestion_shouldReturnError_whenQuestionTypeIsZeroValue(t *testing.T) {
	t.Parallel()

	// given
	id, _, content, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion(id, question.Type{}, content, orderIndex, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewQuestion_shouldReturnError_whenContentIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	id, qt, _, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion(id, qt, "", orderIndex, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewQuestion_shouldReturnError_whenContentExceedsMaxLength(t *testing.T) {
	t.Parallel()

	// given
	id, qt, _, orderIndex, createdAt, updatedAt := validQuestionArgs()
	longContent := strings.Repeat("a", 10001)

	// when
	_, err := question.NewQuestion(id, qt, longContent, orderIndex, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_NewQuestion_shouldReturnError_whenOrderIndexIsNegative(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, _, createdAt, updatedAt := validQuestionArgs()

	// when
	_, err := question.NewQuestion(id, qt, content, -1, createdAt, updatedAt)

	// then
	require.Error(t, err)
}

func Test_ReconstructQuestion_shouldReturnQuestion_withoutValidation(t *testing.T) {
	t.Parallel()

	// given
	id, qt, content, orderIndex, createdAt, updatedAt := validQuestionArgs()

	// when
	q := question.ReconstructQuestion(id, qt, content, orderIndex, createdAt, updatedAt)

	// then
	assert.Equal(t, id, q.ID())
	assert.Equal(t, content, q.Content())
}
