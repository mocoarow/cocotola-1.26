//go:build small

package seed_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

type orderingSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type orderingContent struct {
	Source struct {
		Text string `json:"text"`
		Lang string `json:"lang"`
	} `json:"source"`
	Target struct {
		Lang     string            `json:"lang"`
		Segments []orderingSegment `json:"segments"`
	} `json:"target"`
	Explanation string `json:"explanation"`
}

func parseOrderingContent(t *testing.T, content string) orderingContent {
	t.Helper()
	var c orderingContent
	require.NoError(t, json.Unmarshal([]byte(content), &c))
	return c
}

func orderingEntry() seed.CSVWorkbookEntry {
	return seed.CSVWorkbookEntry{
		SeedKey:     "cefr-b1-ordering-test",
		Title:       "CEFR B1 ordering",
		Description: "desc",
		Language:    "ja",
		Format:      "generated-ordering",
		SourceLang:  "ja",
		TargetLang:  "en",
		GCSObject:   testCSVObject,
	}
}

const (
	orderingHeader = "id,srcText,segments,pattern,patternName,patternPoint,tags,sourceType\n"
	// orderingSegmentsCell is a CSV-quoted segments value: chips "I", "have lived", "here" and a fixed ".".
	orderingSegmentsCell = `"[{""type"":""chip"",""text"":""I""},{""type"":""chip"",""text"":""have lived""},{""type"":""chip"",""text"":""here""},{""type"":""fixed"",""text"":"".""}]"`
)

// orderingRow builds a data row whose other columns are valid.
func orderingRow(id, srcText, segments, pattern, patternName, patternPoint string) string {
	return strings.Join([]string{id, srcText, segments, pattern, patternName, patternPoint, "B1", "generated"}, ",") + "\n"
}

// validOrderingRows returns valid data rows whose ids run from first to last.
func validOrderingRows(first, last int) string {
	var b strings.Builder
	for i := first; i <= last; i++ {
		b.WriteString(orderingRow(fmt.Sprintf("b1-ppc-%d", i), fmt.Sprintf("ここに住んでいる%d", i),
			orderingSegmentsCell, "present-perfect-continuation", "現在完了継続", "継続を表す"))
	}
	return b.String()
}

func loadSingleOrderingSeed(ctx context.Context, t *testing.T, csv string) (seed.CSVWorkbookSeed, error) {
	t.Helper()
	reader := NewMockGCSObjectReader(t)
	reader.EXPECT().ReadObject(ctx, testCSVObject).Return([]byte(csv), nil)
	got, err := seed.LoadCSVWorkbookSeed(ctx, reader, orderingEntry())
	if err != nil {
		return seed.CSVWorkbookSeed{}, fmt.Errorf("load csv workbook seed: %w", err)
	}
	return got, nil
}

func Test_LoadCSVWorkbookSeed_shouldConvertRowToOrderingQuestion_whenGeneratedOrdering(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := orderingHeader + validOrderingRows(1, 2)

	// when
	got, err := loadSingleOrderingSeed(ctx, t, csv)

	// then
	require.NoError(t, err)
	require.Len(t, got.Seed.Questions, 2)
	q := got.Seed.Questions[1]
	assert.Equal(t, "b1-ppc-2", q.SeedKey)
	assert.Equal(t, "ordering", q.QuestionType)
	assert.Equal(t, int32(2), q.OrderIndex)
}

func Test_LoadCSVWorkbookSeed_shouldCopySourceAndSegmentsIntoContent_whenGeneratedOrdering(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := orderingHeader + validOrderingRows(1, 1)

	// when
	got, err := loadSingleOrderingSeed(ctx, t, csv)

	// then
	require.NoError(t, err)
	require.Len(t, got.Seed.Questions, 1)
	c := parseOrderingContent(t, got.Seed.Questions[0].Content)
	assert.Equal(t, "ここに住んでいる1", c.Source.Text)
	assert.Equal(t, "ja", c.Source.Lang)
	assert.Equal(t, "en", c.Target.Lang)
	assert.Equal(t, []orderingSegment{
		{Type: "chip", Text: "I"},
		{Type: "chip", Text: "have lived"},
		{Type: "chip", Text: "here"},
		{Type: "fixed", Text: "."},
	}, c.Target.Segments)
}

// Test_LoadCSVWorkbookSeed_shouldEmitOrderingContentMatchingQuestionWireContract
// guards the mirrored content structs against drifting from cocotola-question/contentjson.
func Test_LoadCSVWorkbookSeed_shouldEmitOrderingContentMatchingQuestionWireContract(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := orderingHeader + validOrderingRows(1, 1)

	// when
	got, err := loadSingleOrderingSeed(ctx, t, csv)

	// then
	require.NoError(t, err)
	var top map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(got.Seed.Questions[0].Content), &top))
	assert.Equal(t, []string{"explanation", "source", "target"}, sortedKeys(top))

	var src map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(top["source"], &src))
	assert.Equal(t, []string{"lang", "text"}, sortedKeys(src))

	var target map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(top["target"], &target))
	assert.Equal(t, []string{"lang", "segments"}, sortedKeys(target))

	var segments []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(target["segments"], &segments))
	require.NotEmpty(t, segments)
	assert.Equal(t, []string{"text", "type"}, sortedKeys(segments[0]))
}

func Test_LoadCSVWorkbookSeed_shouldTagLevelGrammarAndSource_whenGeneratedOrdering(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := orderingHeader + validOrderingRows(1, 1)

	// when
	got, err := loadSingleOrderingSeed(ctx, t, csv)

	// then
	require.NoError(t, err)
	require.Len(t, got.Seed.Questions, 1)
	assert.Equal(t,
		[]string{"level:b1", "grammar:present-perfect-continuation", "source:generated"},
		got.Seed.Questions[0].Tags)
}

func Test_LoadCSVWorkbookSeed_shouldLowercaseGrammarTag_whenPatternHasUppercase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// given
	csv := orderingHeader + orderingRow("b1-ppc-1", "ここに住んでいる", orderingSegmentsCell, "Present-Perfect", "", "")

	// when
	got, err := loadSingleOrderingSeed(ctx, t, csv)

	// then
	require.NoError(t, err)
	require.Len(t, got.Seed.Questions, 1)
	assert.Contains(t, got.Seed.Questions[0].Tags, "grammar:present-perfect")
}

func Test_LoadCSVWorkbookSeed_shouldBuildExplanationFromPattern_whenGeneratedOrdering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		patternName  string
		patternPoint string
		want         string
	}{
		{name: "name and point", patternName: "現在完了継続", patternPoint: "継続を表す", want: "**現在完了継続** 継続を表す"},
		{name: "name only", patternName: "現在完了継続", patternPoint: "", want: "**現在完了継続**"},
		{name: "point only", patternName: "", patternPoint: "継続を表す", want: "継続を表す"},
		{name: "both empty", patternName: "", patternPoint: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			csv := orderingHeader + orderingRow("b1-ppc-1", "ここに住んでいる", orderingSegmentsCell,
				"present-perfect-continuation", tt.patternName, tt.patternPoint)

			// when
			got, err := loadSingleOrderingSeed(ctx, t, csv)

			// then
			require.NoError(t, err)
			require.Len(t, got.Seed.Questions, 1)
			assert.Equal(t, tt.want, parseOrderingContent(t, got.Seed.Questions[0].Content).Explanation)
		})
	}
}

func Test_LoadCSVWorkbookSeed_shouldSkipOrderingRow_whenInvalid(t *testing.T) {
	t.Parallel()

	const pattern = "present-perfect-continuation"
	tests := []struct {
		name       string
		invalidRow string
	}{
		{name: "segments not JSON", invalidRow: orderingRow("b1-ppc-20", "文", "I have lived here", pattern, "", "")},
		{name: "segments not an array", invalidRow: orderingRow("b1-ppc-20", "文", `"{""type"":""chip"",""text"":""I""}"`, pattern, "", "")},
		{name: "segments array of strings", invalidRow: orderingRow("b1-ppc-20", "文", `"[""I"",""here""]"`, pattern, "", "")},
		{name: "segments empty array", invalidRow: orderingRow("b1-ppc-20", "文", "[]", pattern, "", "")},
		{name: "segments null", invalidRow: orderingRow("b1-ppc-20", "文", "null", pattern, "", "")},
		{name: "segments empty", invalidRow: orderingRow("b1-ppc-20", "文", "", pattern, "", "")},
		{name: "pattern outside the tag pattern", invalidRow: orderingRow("b1-ppc-20", "文", orderingSegmentsCell, "present perfect", "", "")},
		{name: "pattern empty", invalidRow: orderingRow("b1-ppc-20", "文", orderingSegmentsCell, "", "", "")},
		{name: "empty id", invalidRow: orderingRow("", "文", orderingSegmentsCell, pattern, "", "")},
		{name: "duplicate id", invalidRow: orderingRow("b1-ppc-1", "文", orderingSegmentsCell, pattern, "", "")},
		{name: "empty source text", invalidRow: orderingRow("b1-ppc-20", "", orderingSegmentsCell, pattern, "", "")},
		{name: "too few columns", invalidRow: "b1-ppc-20,文," + orderingSegmentsCell + "," + pattern + "\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			csv := orderingHeader + validOrderingRows(1, 19) + tt.invalidRow

			// when
			got, err := loadSingleOrderingSeed(ctx, t, csv)

			// then
			require.NoError(t, err)
			assert.Len(t, got.Seed.Questions, 19)
		})
	}
}

func Test_LoadCSVWorkbookSeed_shouldReturnErrMissingCSVColumn_whenOrderingRequiredColumnMissing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		header string
	}{
		{name: "id", header: "srcText,segments,pattern\n"},
		{name: "srcText", header: "id,segments,pattern\n"},
		{name: "segments", header: "id,srcText,pattern\n"},
		{name: "pattern", header: "id,srcText,segments\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// given
			csv := tt.header + "a,b,c\n"

			// when
			_, err := loadSingleOrderingSeed(ctx, t, csv)

			// then
			require.ErrorIs(t, err, seed.ErrMissingCSVColumn)
		})
	}
}
