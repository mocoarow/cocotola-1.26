//go:build small

package seed_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

// wordFillContent mirrors the JSON shape the converter emits, used to assert
// individual fields (avoids brittle full-JSON string comparisons).
type wordFillContent struct {
	Source struct {
		Text string `json:"text"`
		Lang string `json:"lang"`
	} `json:"source"`
	Target struct {
		Text string `json:"text"`
		Lang string `json:"lang"`
	} `json:"target"`
	Explanation1       string `json:"explanation1"`
	Explanation2       string `json:"explanation2"`
	AllowPartialCredit bool   `json:"allowPartialCredit"`
}

func parseWordFillContent(t *testing.T, content string) wordFillContent {
	t.Helper()
	var c wordFillContent
	require.NoError(t, json.Unmarshal([]byte(content), &c))
	return c
}

// tatoebaAttr builds the expected attribution string for a ja/en sentence pair
// where both authors are known.
func tatoebaAttr(jaNumber, jaAuthor, enNumber, enAuthor string) string {
	const license = " / Licensed under [CC BY 2.0 FR](https://creativecommons.org/licenses/by/2.0/fr/)"
	line := func(lang, number, author string) string {
		return fmt.Sprintf("Sentence source(%s): Tatoeba [#%s](https://tatoeba.org/en/sentences/show/%s) / Author: [%s](https://tatoeba.org/en/user/profile/%s)%s",
			lang, number, number, author, author, license)
	}
	return line("ja", jaNumber, jaAuthor) + "\n\n" + line("en", enNumber, enAuthor)
}

const (
	testCSVObject  = "cefr.csv"
	testCSVSeedKey = "cefr-b2-wordfill-v1"
)

// wordFillManifest returns a single-entry manifest for the tatoeba-wordfill
// format used across the loader tests.
func wordFillManifest() seed.CSVWorkbookManifest {
	return seed.CSVWorkbookManifest{
		Workbooks: []seed.CSVWorkbookEntry{
			{
				SeedKey:     testCSVSeedKey,
				Title:       "CEFR B2",
				Description: "desc",
				Language:    "ja",
				Format:      "tatoeba-wordfill",
				SourceLang:  "ja",
				TargetLang:  "en",
				GCSObject:   testCSVObject,
			},
		},
	}
}

const wordFillHeader = "id,wordId,srcText,dstText,blankText,tags,blankAnswers,srcSentenceNumber,srcAuthor,dstSentenceNumber,dstAuthor,sourceType\n"

func Test_LoadCSVWorkbookSeeds_shouldConvertSingleBlankRow_whenWordFill(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := wordFillHeader +
		"1,3,給料は君の能力次第だ。,You will be paid according to your ability.,You will be paid according to your ___.,B1,ability,19526,CK,182349,KK,tatoeba\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	seeds, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.NoError(t, err)
	require.Len(t, seeds, 1)
	require.Len(t, seeds[0].Questions, 1)
	q := seeds[0].Questions[0]
	assert.Equal(t, "1", q.SeedKey)
	assert.Equal(t, "word_fill", q.QuestionType)
	assert.Equal(t, int32(1), q.OrderIndex)
	c := parseWordFillContent(t, q.Content)
	assert.Equal(t, "給料は君の能力次第だ。", c.Source.Text)
	assert.Equal(t, "ja", c.Source.Lang)
	assert.Equal(t, "You will be paid according to your {{ability}}.", c.Target.Text)
	assert.Equal(t, "en", c.Target.Lang)
	assert.False(t, c.AllowPartialCredit)
	wantAttr := tatoebaAttr("182349", "KK", "19526", "CK")
	assert.Equal(t, wantAttr, c.Explanation1)
	assert.Equal(t, wantAttr, c.Explanation2)
	assert.Equal(t, []string{"level:b1", "source:tatoeba"}, q.Tags)
}

func Test_LoadCSVWorkbookSeeds_shouldConvertMultiBlankRow_whenWordFill(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := wordFillHeader +
		`5,6,気を楽にして。,"Relax, and above all, don't panic.","Relax, and ___ ___, don't panic.",B1,"above,all",1,CK,2,KK,tatoeba` + "\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	seeds, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.NoError(t, err)
	require.Len(t, seeds[0].Questions, 1)
	c := parseWordFillContent(t, seeds[0].Questions[0].Content)
	assert.Equal(t, "Relax, and {{above}} {{all}}, don't panic.", c.Target.Text)
	assert.True(t, c.AllowPartialCredit)
	wantAttr := tatoebaAttr("2", "KK", "1", "CK")
	assert.Equal(t, wantAttr, c.Explanation1)
	assert.Equal(t, wantAttr, c.Explanation2)
}

func Test_LoadCSVWorkbookSeeds_shouldOmitAuthorSegment_whenAuthorEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: the ja sentence (dst* columns) has no author
	csv := wordFillHeader +
		"7,3,日本語,English ability.,English ___.,B1,ability,19526,CK,182349,,tatoeba\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	seeds, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.NoError(t, err)
	c := parseWordFillContent(t, seeds[0].Questions[0].Content)
	want := "Sentence source(ja): Tatoeba [#182349](https://tatoeba.org/en/sentences/show/182349) / Licensed under [CC BY 2.0 FR](https://creativecommons.org/licenses/by/2.0/fr/)" +
		"\n\n" +
		"Sentence source(en): Tatoeba [#19526](https://tatoeba.org/en/sentences/show/19526) / Author: [CK](https://tatoeba.org/en/user/profile/CK) / Licensed under [CC BY 2.0 FR](https://creativecommons.org/licenses/by/2.0/fr/)"
	assert.Equal(t, want, c.Explanation1)
}

func Test_LoadCSVWorkbookSeeds_shouldSetWorkbookMetadataFromManifest_notFromCSV(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := wordFillHeader +
		"1,3,日本語,English ability.,English ___.,B1,ability,1,CK,2,KK,tatoeba\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	seeds, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.NoError(t, err)
	assert.Equal(t, testCSVSeedKey, seeds[0].SeedKey)
	assert.Equal(t, "CEFR B2", seeds[0].Title)
	assert.Equal(t, "desc", seeds[0].Description)
	assert.Equal(t, "ja", seeds[0].Language)
}

// sortedKeys returns the sorted top-level keys of a decoded JSON object, so key
// presence can be asserted independent of marshalling order.
func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Test_LoadCSVWorkbookSeeds_shouldEmitContentMatchingQuestionWireContract pins
// the JSON shape of the generated word_fill content. This package deliberately
// does not import cocotola-question's domain (the content structs are mirrored
// in csv_converter.go), so this test guards against that mirror silently
// drifting from the keys cocotola-question's WordFillContent accepts.
func Test_LoadCSVWorkbookSeeds_shouldEmitContentMatchingQuestionWireContract(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := wordFillHeader +
		"1,3,給料は君の能力次第だ。,You will be paid according to your ability.,You will be paid according to your ___.,B1,ability,19526,CK,182349,KK,tatoeba\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	seeds, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.NoError(t, err)
	var top map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(seeds[0].Questions[0].Content), &top))
	assert.Equal(t,
		[]string{"allowPartialCredit", "explanation1", "explanation2", "source", "target"},
		sortedKeys(top))

	var src map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(top["source"], &src))
	assert.Equal(t, []string{"lang", "text"}, sortedKeys(src))
}

func Test_LoadCSVWorkbookSeeds_shouldReturnError_whenBlankCountMismatchesAnswers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: one blank but two answers
	csv := wordFillHeader +
		`1,3,日本語,English.,English ___.,B1,"one,two",1,CK,2,KK,tatoeba` + "\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	_, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.ErrorIs(t, err, seed.ErrInvalidCSVRow)
}

func Test_LoadCSVWorkbookSeeds_shouldReturnError_whenRequiredColumnMissing(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: header without blankAnswers
	csv := "id,srcText,blankText\n1,日本語,English ___.\n"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)

	// when
	_, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.ErrorIs(t, err, seed.ErrMissingCSVColumn)
}

func Test_LoadCSVWorkbookSeeds_shouldReturnError_whenFormatUnsupported(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	manifest := wordFillManifest()
	manifest.Workbooks[0].Format = "unknown-format"
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(wordFillHeader), nil)

	// when
	_, err := seed.LoadCSVWorkbookSeeds(ctx, reader, manifest)

	// then
	require.ErrorIs(t, err, seed.ErrUnsupportedCSVFormat)
}

func Test_LoadCSVWorkbookSeeds_shouldReturnError_whenReaderFails(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	readErr := errors.New("object not found")
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return(nil, readErr)

	// when
	_, err := seed.LoadCSVWorkbookSeeds(ctx, reader, wordFillManifest())

	// then
	require.ErrorIs(t, err, readErr)
}

func Test_LoadCSVWorkbookSeeds_shouldReturnError_whenManifestHasDuplicateSeedKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given: two entries share a seedKey
	manifest := seed.CSVWorkbookManifest{
		Workbooks: []seed.CSVWorkbookEntry{
			{SeedKey: "dup", Title: "A", Format: "tatoeba-wordfill", SourceLang: "ja", TargetLang: "en", GCSObject: "a.csv"},
			{SeedKey: "dup", Title: "B", Format: "tatoeba-wordfill", SourceLang: "ja", TargetLang: "en", GCSObject: "b.csv"},
		},
	}
	reader := NewMockGCSObjectReader(t)

	// when
	_, err := seed.LoadCSVWorkbookSeeds(ctx, reader, manifest)

	// then
	require.ErrorIs(t, err, seed.ErrInvalidManifest)
}
