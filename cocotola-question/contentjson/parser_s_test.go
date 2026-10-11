package contentjson_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/contentjson"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

const (
	fixtureWordFillJSON = `{"source":{"text":"ゴミを捨てる","lang":"ja"},"target":{"text":"{{throw}} it {{away}}","lang":"en"},` +
		`"explanation":"e0","explanation1":"e1","explanation2":"e2","allowPartialCredit":true}`
	fixtureMultipleChoiceJSON = `{"questionText":"Pick the oceans","explanation":"ex",` +
		`"choices":[{"id":"c1","text":"Pacific","isCorrect":true},{"id":"c2","text":"Mt Fuji","isCorrect":false}],` +
		`"displayCount":2,"showCorrectCount":true,"shuffleChoices":true,"allowPartialCredit":true}`
)

func Test_Parser_ParseWordFill_shouldMapAllFields_whenJSONIsValid(t *testing.T) {
	t.Parallel()

	// given
	parser := contentjson.Parser{}

	// when
	got, err := parser.ParseWordFill(fixtureWordFillJSON)

	// then
	require.NoError(t, err)
	assert.Equal(t, &question.WordFillContent{
		Source:             question.TextWithLang{Text: "ゴミを捨てる", Lang: "ja"},
		Target:             question.TextWithLang{Text: "{{throw}} it {{away}}", Lang: "en"},
		Explanation:        "e0",
		Explanation1:       "e1",
		Explanation2:       "e2",
		AllowPartialCredit: true,
	}, got)
}

func Test_Parser_ParseWordFill_shouldReturnErrInvalidArgument_whenJSONIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	parser := contentjson.Parser{}

	// when
	_, err := parser.ParseWordFill("not json")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}

func Test_Parser_ParseMultipleChoice_shouldMapAllFields_whenJSONIsValid(t *testing.T) {
	t.Parallel()

	// given
	parser := contentjson.Parser{}

	// when
	got, err := parser.ParseMultipleChoice(fixtureMultipleChoiceJSON)

	// then
	require.NoError(t, err)
	assert.Equal(t, &question.MultipleChoiceContent{
		QuestionText: "Pick the oceans",
		Explanation:  "ex",
		Choices: []question.Choice{
			{ID: "c1", Text: "Pacific", IsCorrect: true},
			{ID: "c2", Text: "Mt Fuji", IsCorrect: false},
		},
		DisplayCount:       2,
		ShowCorrectCount:   true,
		ShuffleChoices:     true,
		AllowPartialCredit: true,
	}, got)
}

func Test_Parser_ParseMultipleChoice_shouldReturnErrInvalidArgument_whenJSONIsInvalid(t *testing.T) {
	t.Parallel()

	// given
	parser := contentjson.Parser{}

	// when
	_, err := parser.ParseMultipleChoice("not json")

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
