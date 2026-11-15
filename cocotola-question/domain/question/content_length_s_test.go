package question_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

const multibyteChar = "あ"

func multibyteText(n int) string {
	return strings.Repeat(multibyteChar, n)
}

func wordFillWithTarget(text string) *question.WordFillContent {
	c := validWordFillContent()
	c.Target.Text = text
	return c
}

type wordFillLengthCase struct {
	name    string
	content *question.WordFillContent
}

func wordFillLengthCases(textLength int, explanationLength int) []wordFillLengthCase {
	source := validWordFillContent()
	source.Source.Text = multibyteText(textLength)
	explanation := validWordFillContent()
	explanation.Explanation = multibyteText(explanationLength)
	explanation1 := validWordFillContent()
	explanation1.Explanation1 = multibyteText(explanationLength)
	explanation2 := validWordFillContent()
	explanation2.Explanation2 = multibyteText(explanationLength)
	const blank = "{{a}}"
	return []wordFillLengthCase{
		{name: "source text", content: source},
		{name: "target text", content: wordFillWithTarget(blank + multibyteText(textLength-len(blank)))},
		{name: "explanation", content: explanation},
		{name: "explanation1", content: explanation1},
		{name: "explanation2", content: explanation2},
	}
}

func Test_WordFillContent_Validate_shouldSucceed_whenMultibyteTextIsAtMaxLength(t *testing.T) {
	t.Parallel()

	tests := wordFillLengthCases(question.MaxQuestionTextLength, question.MaxExplanationLength)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := tt.content.Validate()

			// then
			require.NoError(t, err)
		})
	}
}

func Test_WordFillContent_Validate_shouldReturnError_whenMultibyteTextExceedsMaxLength(t *testing.T) {
	t.Parallel()

	tests := wordFillLengthCases(question.MaxQuestionTextLength+1, question.MaxExplanationLength+1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := tt.content.Validate()

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

type multipleChoiceLengthCase struct {
	name    string
	content *question.MultipleChoiceContent
}

func multipleChoiceLengthCases(textLength int, explanationLength int, choiceLength int) []multipleChoiceLengthCase {
	questionText := validMultipleChoiceContent()
	questionText.QuestionText = multibyteText(textLength)
	explanation := validMultipleChoiceContent()
	explanation.Explanation = multibyteText(explanationLength)
	choice := validMultipleChoiceContent()
	choice.Choices[0].Text = multibyteText(choiceLength)
	return []multipleChoiceLengthCase{
		{name: "question text", content: questionText},
		{name: "explanation", content: explanation},
		{name: "choice text", content: choice},
	}
}

func Test_MultipleChoiceContent_Validate_shouldSucceed_whenMultibyteTextIsAtMaxLength(t *testing.T) {
	t.Parallel()

	tests := multipleChoiceLengthCases(question.MaxQuestionTextLength, question.MaxExplanationLength, question.MaxChoiceTextLength)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := tt.content.Validate()

			// then
			require.NoError(t, err)
		})
	}
}

func Test_MultipleChoiceContent_Validate_shouldReturnError_whenMultibyteTextExceedsMaxLength(t *testing.T) {
	t.Parallel()

	tests := multipleChoiceLengthCases(question.MaxQuestionTextLength+1, question.MaxExplanationLength+1, question.MaxChoiceTextLength+1)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := tt.content.Validate()

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}
