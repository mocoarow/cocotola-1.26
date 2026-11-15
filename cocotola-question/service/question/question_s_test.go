//go:build small

package question_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/service/question"
)

const (
	maxContentLength = 10000
	maxTags          = 20
	maxTagLength     = 100
)

func repeatTag(n int) []string {
	tags := make([]string, n)
	for i := range tags {
		tags[i] = "tag"
	}
	return tags
}

type contentFields struct {
	name       string
	content    string
	tags       []string
	orderIndex int
}

func contentBoundaryCases() []contentFields {
	return []contentFields{
		{name: "content at max length", content: strings.Repeat("a", maxContentLength), tags: []string{"t"}, orderIndex: 1},
		{name: "tags at max count", content: "c", tags: repeatTag(maxTags), orderIndex: 1},
		{name: "tag at max length", content: "c", tags: []string{strings.Repeat("a", maxTagLength)}, orderIndex: 1},
		{name: "tags nil", content: "c", tags: nil, orderIndex: 1},
		{name: "orderIndex zero", content: "c", tags: []string{"t"}, orderIndex: 0},
	}
}

func contentInvalidCases() []contentFields {
	return []contentFields{
		{name: "content empty", content: "", tags: []string{"t"}, orderIndex: 1},
		{name: "content over max length", content: strings.Repeat("a", maxContentLength+1), tags: []string{"t"}, orderIndex: 1},
		{name: "tags over max count", content: "c", tags: repeatTag(maxTags + 1), orderIndex: 1},
		{name: "tag over max length", content: "c", tags: []string{strings.Repeat("a", maxTagLength+1)}, orderIndex: 1},
		{name: "orderIndex negative", content: "c", tags: []string{"t"}, orderIndex: -1},
	}
}

type questionIDs struct {
	name           string
	operatorID     string
	organizationID string
	workbookID     string
	questionID     string
}

func emptyQuestionIDCases() []questionIDs {
	return []questionIDs{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", workbookID: "wb1", questionID: "q1"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", workbookID: "wb1", questionID: "q1"},
		{name: "workbookID empty", operatorID: "op1", organizationID: "org1", workbookID: "", questionID: "q1"},
		{name: "questionID empty", operatorID: "op1", organizationID: "org1", workbookID: "wb1", questionID: ""},
	}
}

func Test_NewAddQuestionInput_shouldReturnInput_whenFieldsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := question.NewAddQuestionInput("op1", "org1", "wb1", "word_fill", "content1", []string{"tag1", "tag2"}, 3)

	// then
	require.NoError(t, err)
	assert.Equal(t, question.AddQuestionInput{
		OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1", QuestionType: "word_fill",
		Content: "content1", Tags: []string{"tag1", "tag2"}, OrderIndex: 3,
	}, *input)
}

func Test_NewAddQuestionInput_shouldReturnInput_whenFieldsAreAtBoundary(t *testing.T) {
	t.Parallel()

	tests := contentBoundaryCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewAddQuestionInput("op1", "org1", "wb1", "word_fill", tt.content, tt.tags, tt.orderIndex)

			// then
			require.NoError(t, err)
		})
	}
}

func Test_NewAddQuestionInput_shouldReturnErrInvalidArgument_whenContentFieldIsInvalid(t *testing.T) {
	t.Parallel()

	tests := contentInvalidCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewAddQuestionInput("op1", "org1", "wb1", "word_fill", tt.content, tt.tags, tt.orderIndex)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewAddQuestionInput_shouldReturnErrInvalidArgument_whenRequiredStringIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
		workbookID     string
		questionType   string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", workbookID: "wb1", questionType: "word_fill"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", workbookID: "wb1", questionType: "word_fill"},
		{name: "workbookID empty", operatorID: "op1", organizationID: "org1", workbookID: "", questionType: "word_fill"},
		{name: "questionType empty", operatorID: "op1", organizationID: "org1", workbookID: "wb1", questionType: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewAddQuestionInput(tt.operatorID, tt.organizationID, tt.workbookID, tt.questionType, "content1", []string{"tag1"}, 1)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUpdateQuestionInput_shouldReturnInput_whenFieldsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := question.NewUpdateQuestionInput("op1", "org1", "wb1", "q1", "content1", []string{"tag1", "tag2"}, 3)

	// then
	require.NoError(t, err)
	assert.Equal(t, question.UpdateQuestionInput{
		OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1", QuestionID: "q1",
		Content: "content1", Tags: []string{"tag1", "tag2"}, OrderIndex: 3,
	}, *input)
}

func Test_NewUpdateQuestionInput_shouldReturnInput_whenFieldsAreAtBoundary(t *testing.T) {
	t.Parallel()

	tests := contentBoundaryCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewUpdateQuestionInput("op1", "org1", "wb1", "q1", tt.content, tt.tags, tt.orderIndex)

			// then
			require.NoError(t, err)
		})
	}
}

func Test_NewUpdateQuestionInput_shouldReturnErrInvalidArgument_whenContentFieldIsInvalid(t *testing.T) {
	t.Parallel()

	tests := contentInvalidCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewUpdateQuestionInput("op1", "org1", "wb1", "q1", tt.content, tt.tags, tt.orderIndex)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewUpdateQuestionInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := emptyQuestionIDCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewUpdateQuestionInput(tt.operatorID, tt.organizationID, tt.workbookID, tt.questionID, "content1", []string{"tag1"}, 1)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewGetQuestionInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := question.NewGetQuestionInput("op1", "org1", "wb1", "q1")

	// then
	require.NoError(t, err)
	assert.Equal(t, question.GetQuestionInput{OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1", QuestionID: "q1"}, *input)
}

func Test_NewGetQuestionInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := emptyQuestionIDCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewGetQuestionInput(tt.operatorID, tt.organizationID, tt.workbookID, tt.questionID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewListQuestionsInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := question.NewListQuestionsInput("op1", "org1", "wb1")

	// then
	require.NoError(t, err)
	assert.Equal(t, question.ListQuestionsInput{OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1"}, *input)
}

func Test_NewListQuestionsInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		operatorID     string
		organizationID string
		workbookID     string
	}{
		{name: "operatorID empty", operatorID: "", organizationID: "org1", workbookID: "wb1"},
		{name: "organizationID empty", operatorID: "op1", organizationID: "", workbookID: "wb1"},
		{name: "workbookID empty", operatorID: "op1", organizationID: "org1", workbookID: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewListQuestionsInput(tt.operatorID, tt.organizationID, tt.workbookID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_NewDeleteQuestionInput_shouldReturnInput_whenIDsAreSet(t *testing.T) {
	t.Parallel()

	// given, when
	input, err := question.NewDeleteQuestionInput("op1", "org1", "wb1", "q1")

	// then
	require.NoError(t, err)
	assert.Equal(t, question.DeleteQuestionInput{OperatorID: "op1", OrganizationID: "org1", WorkbookID: "wb1", QuestionID: "q1"}, *input)
}

func Test_NewDeleteQuestionInput_shouldReturnErrInvalidArgument_whenIDIsEmpty(t *testing.T) {
	t.Parallel()

	tests := emptyQuestionIDCases()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given, when
			_, err := question.NewDeleteQuestionInput(tt.operatorID, tt.organizationID, tt.workbookID, tt.questionID)

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
