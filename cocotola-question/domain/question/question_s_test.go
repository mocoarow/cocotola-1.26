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
