//go:build small

package seed

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// converterTestIdx is the column index for the standard test CSV header.
var converterTestIdx = indexColumns([]string{
	"id", "wordId", "srcText", "dstText", "blankText", "tags",
	"blankAnswers", "srcSentenceNumber", "srcAuthor",
	"dstSentenceNumber", "dstAuthor", "sourceType",
})

func makeConverterRow(srcText, blankText, blankAnswers, srcNum, srcAuthor, dstNum, dstAuthor, tags, sourceType string) []string {
	return []string{
		"", "", srcText, "", blankText, tags,
		blankAnswers, srcNum, srcAuthor, dstNum, dstAuthor, sourceType,
	}
}

type converterWordFillContent struct {
	Source struct {
		Text string `json:"text"`
		Lang string `json:"lang"`
	} `json:"source"`
	Target struct {
		Text string `json:"text"`
		Lang string `json:"lang"`
	} `json:"target"`
	AllowPartialCredit bool `json:"allowPartialCredit"`
}

func parseConverterContent(t *testing.T, content string) converterWordFillContent {
	t.Helper()
	var c converterWordFillContent
	require.NoError(t, json.Unmarshal([]byte(content), &c))
	return c
}

func Test_convertWordFillRow_shouldReturnSeed_whenSingleBlank(t *testing.T) {
	t.Parallel()

	// given
	row := makeConverterRow("給料は君の能力次第だ。", "You will be paid according to your ___.", "ability", "19526", "CK", "182349", "KK", "B1", "tatoeba")

	// when
	got, err := convertWordFillRow(row, converterTestIdx, 2, "1", "ja", "en", 1)

	// then
	require.NoError(t, err)
	assert.Equal(t, "1", got.SeedKey)
	assert.Equal(t, questionTypeWordFill, got.QuestionType)
	assert.Equal(t, int32(1), got.OrderIndex)
	assert.Equal(t, []string{"level:b1", "source:tatoeba"}, got.Tags)
	c := parseConverterContent(t, got.Content)
	assert.Equal(t, "給料は君の能力次第だ。", c.Source.Text)
	assert.Equal(t, "ja", c.Source.Lang)
	assert.Equal(t, "You will be paid according to your {{ability}}.", c.Target.Text)
	assert.Equal(t, "en", c.Target.Lang)
	assert.False(t, c.AllowPartialCredit)
}

func Test_convertWordFillRow_shouldSetAllowPartialCredit_whenMultipleBlanks(t *testing.T) {
	t.Parallel()

	// given
	row := makeConverterRow("気を楽にして。", "Relax, and ___ ___, don't panic.", "above,all", "1", "CK", "2", "KK", "B1", "tatoeba")

	// when
	got, err := convertWordFillRow(row, converterTestIdx, 2, "5", "ja", "en", 1)

	// then
	require.NoError(t, err)
	c := parseConverterContent(t, got.Content)
	assert.Equal(t, "Relax, and {{above}} {{all}}, don't panic.", c.Target.Text)
	assert.True(t, c.AllowPartialCredit)
}

func Test_convertWordFillRow_shouldReturnError_whenBlankCountMismatchesAnswers(t *testing.T) {
	t.Parallel()

	// given: one blank but two answers
	row := makeConverterRow("日本語", "English ___.", "one,two", "1", "CK", "2", "KK", "B1", "tatoeba")

	// when
	_, err := convertWordFillRow(row, converterTestIdx, 2, "1", "ja", "en", 1)

	// then
	require.ErrorIs(t, err, ErrInvalidCSVRow)
}

func Test_convertWordFillRow_shouldReturnError_whenBlankTextHasNoPlaceholder(t *testing.T) {
	t.Parallel()

	// given: blankText contains no "___"
	row := makeConverterRow("日本語", "English sentence.", "ability", "1", "CK", "2", "KK", "B1", "tatoeba")

	// when
	_, err := convertWordFillRow(row, converterTestIdx, 2, "1", "ja", "en", 1)

	// then
	require.ErrorIs(t, err, ErrInvalidCSVRow)
}
