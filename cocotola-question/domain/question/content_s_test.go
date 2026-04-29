package question_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}

func validWordFillJSON(t *testing.T) string {
	t.Helper()
	return mustMarshal(t, question.WordFillContent{
		Source:             question.TextWithLang{Text: "ゴミを捨てる", Lang: "ja"},
		Target:             question.TextWithLang{Text: "{{throw}} it {{away}}", Lang: "en"},
		Explanation:        "throw away は句動詞です。",
		AllowPartialCredit: true,
	})
}

func validMultipleChoiceJSON(t *testing.T) string {
	t.Helper()
	return mustMarshal(t, question.MultipleChoiceContent{
		QuestionText: "日本の首都はどこですか？",
		Explanation:  "東京は日本の首都です。",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "東京", IsCorrect: true},
			{ID: "c2", Text: "大阪", IsCorrect: false},
			{ID: "c3", Text: "京都", IsCorrect: false},
			{ID: "c4", Text: "名古屋", IsCorrect: false},
		},
		DisplayCount:       4,
		ShowCorrectCount:   true,
		ShuffleChoices:     true,
		AllowPartialCredit: false,
	})
}

func Test_ValidateContent_shouldSucceed_whenWordFillIsValid(t *testing.T) {
	t.Parallel()

	// when
	err := question.ValidateContent(question.TypeWordFill(), validWordFillJSON(t))

	// then
	require.NoError(t, err)
}

func Test_ValidateContent_shouldSucceed_whenSingleWordFill(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "りんご", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{apple}}", Lang: "en"},
	}
	// when
	err := question.ValidateContent(question.TypeWordFill(), mustMarshal(t, c))

	// then
	require.NoError(t, err)
}

func Test_ValidateContent_shouldReturnError_whenWordFillHasNoBlanks(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "ゴミを捨てる", Lang: "ja"},
		Target: question.TextWithLang{Text: "throw it away", Lang: "en"},
	}
	// when
	err := question.ValidateContent(question.TypeWordFill(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenWordFillSameLang(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "hello", Lang: "en"},
		Target: question.TextWithLang{Text: "{{world}}", Lang: "en"},
	}
	// when
	err := question.ValidateContent(question.TypeWordFill(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenWordFillHasHTMLInSource(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "<script>alert('xss')</script>", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{hello}}", Lang: "en"},
	}
	// when
	err := question.ValidateContent(question.TypeWordFill(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenWordFillHasEmptyBlank(t *testing.T) {
	t.Parallel()

	// given
	c := question.WordFillContent{
		Source: question.TextWithLang{Text: "テスト", Lang: "ja"},
		Target: question.TextWithLang{Text: "{{ }}", Lang: "en"},
	}
	// when
	err := question.ValidateContent(question.TypeWordFill(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldSucceed_whenMultipleChoiceIsValid(t *testing.T) {
	t.Parallel()

	// when
	err := question.ValidateContent(question.TypeMultipleChoice(), validMultipleChoiceJSON(t))

	// then
	require.NoError(t, err)
}

func Test_ValidateContent_shouldReturnError_whenMultipleChoiceHasNoCorrect(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "選択肢A", IsCorrect: false},
			{ID: "c2", Text: "選択肢B", IsCorrect: false},
		},
		DisplayCount: 2,
	}
	// when
	err := question.ValidateContent(question.TypeMultipleChoice(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenMultipleChoiceDisplayCountExceedsChoices(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
			{ID: "c2", Text: "選択肢B", IsCorrect: false},
		},
		DisplayCount: 5,
	}
	// when
	err := question.ValidateContent(question.TypeMultipleChoice(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenDisplayCountLessThanCorrectCount(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
			{ID: "c2", Text: "選択肢B", IsCorrect: true},
			{ID: "c3", Text: "選択肢C", IsCorrect: true},
			{ID: "c4", Text: "選択肢D", IsCorrect: false},
		},
		DisplayCount: 2,
	}
	// when
	err := question.ValidateContent(question.TypeMultipleChoice(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenMultipleChoiceHasDuplicateIDs(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "テスト問題",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
			{ID: "c1", Text: "選択肢B", IsCorrect: false},
		},
		DisplayCount: 2,
	}
	// when
	err := question.ValidateContent(question.TypeMultipleChoice(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_ValidateContent_shouldReturnError_whenMultipleChoiceHasHTMLInQuestion(t *testing.T) {
	t.Parallel()

	// given
	c := question.MultipleChoiceContent{
		QuestionText: "<b>太字</b>の問題",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "選択肢A", IsCorrect: true},
		},
		DisplayCount: 1,
	}
	// when
	err := question.ValidateContent(question.TypeMultipleChoice(), mustMarshal(t, c))

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func fixtureMultipleChoiceContent() *question.MultipleChoiceContent {
	return &question.MultipleChoiceContent{
		QuestionText: "Pick the oceans",
		Choices: []question.ChoiceJSON{
			{ID: "c1", Text: "Pacific", IsCorrect: true},
			{ID: "c2", Text: "Atlantic", IsCorrect: true},
			{ID: "c3", Text: "Mt Fuji", IsCorrect: false},
			{ID: "c4", Text: "Amazon River", IsCorrect: false},
		},
		DisplayCount: 4,
	}
}

func Test_ParseMultipleChoiceContent_shouldReturnContent_whenJSONIsValid(t *testing.T) {
	t.Parallel()

	// given
	raw := mustMarshal(t, fixtureMultipleChoiceContent())

	// when
	got, err := question.ParseMultipleChoiceContent(raw)

	// then
	require.NoError(t, err)
	require.Equal(t, "Pick the oceans", got.QuestionText)
	require.Len(t, got.Choices, 4)
}

func Test_ParseMultipleChoiceContent_shouldReturnError_whenJSONIsInvalid(t *testing.T) {
	t.Parallel()

	// when
	_, err := question.ParseMultipleChoiceContent("not json")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
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
