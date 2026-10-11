package question_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func validWordFillContent() *question.WordFillContent {
	return &question.WordFillContent{
		Source:             question.TextWithLang{Text: "ゴミを捨てる", Lang: "ja"},
		Target:             question.TextWithLang{Text: "{{throw}} it {{away}}", Lang: "en"},
		Explanation:        "throw away は句動詞です。",
		AllowPartialCredit: true,
	}
}

func validMultipleChoiceContent() *question.MultipleChoiceContent {
	return &question.MultipleChoiceContent{
		QuestionText: "日本の首都はどこですか？",
		Explanation:  "東京は日本の首都です。",
		Choices: []question.Choice{
			{ID: "c1", Text: "東京", IsCorrect: true},
			{ID: "c2", Text: "大阪", IsCorrect: false},
			{ID: "c3", Text: "京都", IsCorrect: false},
			{ID: "c4", Text: "名古屋", IsCorrect: false},
		},
		DisplayCount:       4,
		ShowCorrectCount:   true,
		ShuffleChoices:     true,
		AllowPartialCredit: false,
	}
}

func Test_WordFillContent_Validate_shouldSucceed_whenContentIsValid(t *testing.T) {
	t.Parallel()

	// given
	c := validWordFillContent()

	// when
	err := c.Validate()

	// then
	require.NoError(t, err)
}

func Test_WordFillContent_Validate_shouldSucceed_whenSingleBlank(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんご", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{apple}}", Lang: "en"},
	}

	// when
	err := c.Validate()

	// then
	require.NoError(t, err)
}

func Test_WordFillContent_Validate_shouldSucceed_whenExplanation1And2AreSet(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source:       question.TextWithLang{Text: "りんご", Lang: "ja"},
		Target:       question.TextWithLang{Text: "{{apple}}", Lang: "en"},
		Explanation1: "Sentence source(ja): Tatoeba [#1](https://tatoeba.org/en/sentences/show/1)",
		Explanation2: "Sentence source(en): Tatoeba [#2](https://tatoeba.org/en/sentences/show/2)",
	}

	// when
	err := c.Validate()

	// then
	require.NoError(t, err)
}

func Test_WordFillContent_Validate_shouldReturnError_whenExplanation1HasHTML(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source:       question.TextWithLang{Text: "りんご", Lang: "ja"},
		Target:       question.TextWithLang{Text: "{{apple}}", Lang: "en"},
		Explanation1: "<b>not allowed</b>",
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_WordFillContent_Validate_shouldReturnError_whenTargetHasNoBlanks(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "ゴミを捨てる", Lang: "ja"},
		Target: question.TextWithLang{Text: "throw it away", Lang: "en"},
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_WordFillContent_Validate_shouldReturnError_whenSourceAndTargetLangAreSame(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "hello", Lang: "en"},
		Target: question.TextWithLang{Text: "{{world}}", Lang: "en"},
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_WordFillContent_Validate_shouldReturnError_whenSourceHasHTML(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "<script>alert('xss')</script>", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{hello}}", Lang: "en"},
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_WordFillContent_Validate_shouldReturnError_whenBlankIsEmpty(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "テスト", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{ }}", Lang: "en"},
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_MultipleChoiceContent_Validate_shouldSucceed_whenContentIsValid(t *testing.T) {
	t.Parallel()

	// given
	c := validMultipleChoiceContent()

	// when
	err := c.Validate()

	// then
	require.NoError(t, err)
}

func Test_MultipleChoiceContent_Validate_shouldReturnError_whenNoChoiceIsCorrect(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.Choice{
			{ID: "c1", Text: "選択肢A", IsCorrect: false},
			{ID: "c2", Text: "選択肢B", IsCorrect: false},
		},
		DisplayCount: 2,
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_MultipleChoiceContent_Validate_shouldReturnError_whenDisplayCountExceedsChoices(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.Choice{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
			{ID: "c2", Text: "選択肢B", IsCorrect: false},
		},
		DisplayCount: 5,
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_MultipleChoiceContent_Validate_shouldReturnError_whenDisplayCountIsLessThanCorrectCount(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.Choice{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
			{ID: "c2", Text: "選択肢B", IsCorrect: true},
			{ID: "c3", Text: "選択肢C", IsCorrect: true},
			{ID: "c4", Text: "選択肢D", IsCorrect: false},
		},
		DisplayCount: 2,
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_MultipleChoiceContent_Validate_shouldReturnError_whenChoiceIDsAreDuplicated(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.Choice{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
			{ID: "c1", Text: "選択肢B", IsCorrect: false},
		},
		DisplayCount: 2,
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_MultipleChoiceContent_Validate_shouldReturnError_whenQuestionTextHasHTML(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "<b>太字</b>の問題",
		Choices: []question.Choice{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
		},
		DisplayCount: 1,
	}

	// when
	err := c.Validate()

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func fixtureMultipleChoiceContent() *question.MultipleChoiceContent {
	return &question.MultipleChoiceContent{
		QuestionText: "Pick the oceans",
		Choices: []question.Choice{
			{ID: "c1", Text: "Pacific", IsCorrect: true},
			{ID: "c2", Text: "Atlantic", IsCorrect: true},
			{ID: "c3", Text: "Mt Fuji", IsCorrect: false},
			{ID: "c4", Text: "Amazon River", IsCorrect: false},
		},
		DisplayCount: 4,
	}
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldReturnTrue_whenAllAndOnlyCorrectIDsAreSelected(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	ok, err := c.EvaluateAnswer([]string{"c1", "c2"})

	// then
	require.NoError(t, err)
	require.True(t, ok)
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldIgnoreOrder(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	ok, err := c.EvaluateAnswer([]string{"c2", "c1"})

	// then
	require.NoError(t, err)
	require.True(t, ok)
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldReturnFalse_whenSubsetOfCorrectIsSelected(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	ok, err := c.EvaluateAnswer([]string{"c1"})

	// then
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldReturnFalse_whenSupersetIncludesIncorrect(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	ok, err := c.EvaluateAnswer([]string{"c1", "c2", "c3"})

	// then
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldReturnFalse_whenEmptySelection(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	ok, err := c.EvaluateAnswer([]string{})

	// then
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldReturnFalse_whenOnlyIncorrectsSelected(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	ok, err := c.EvaluateAnswer([]string{"c3", "c4"})

	// then
	require.NoError(t, err)
	require.False(t, ok)
}

func Test_MultipleChoiceContent_EvaluateAnswer_shouldReturnError_whenUnknownChoiceID(t *testing.T) {
	t.Parallel()

	// given
	c := fixtureMultipleChoiceContent()

	// when
	_, err := c.EvaluateAnswer([]string{"c1", "bogus"})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
