package seed

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// formatGeneratedOrdering is the CSV format whose rows hold a Japanese
	// source sentence plus the English sentence split into segments as a JSON
	// array. Each row becomes one ordering question.
	formatGeneratedOrdering = "generated-ordering"

	// questionTypeOrdering mirrors cocotola-question's ordering type value.
	questionTypeOrdering = "ordering"
)

// orderingContentJSON mirrors the ordering JSON of cocotola-question/contentjson
// without importing it, so the seed package has no cocotola-question dependency.
type orderingContentJSON struct {
	Source      textWithLangJSON   `json:"source"`
	Target      orderingTargetJSON `json:"target"`
	Explanation string             `json:"explanation"`
}

type orderingTargetJSON struct {
	Lang     string                `json:"lang"`
	Segments []orderingSegmentJSON `json:"segments"`
}

type orderingSegmentJSON struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// convertOrderingCSV maps each valid data row to an ordering QuestionSeed.
// Segment rules (types, chip count, forbidden characters) are left to
// cocotola-question, so a row breaking them is counted as rejected there.
func convertOrderingCSV(sourceLang, targetLang string, data []byte) (csvConversion, error) {
	return convertCSVRows(data, []string{"id", "srcText", "segments", "pattern"},
		func(row []string, idx map[string]int, id string, orderIndex int32) (QuestionSeed, error) {
			return convertOrderingRow(row, idx, id, sourceLang, targetLang, orderIndex)
		})
}

func convertOrderingRow(row []string, idx map[string]int, id, sourceLang, targetLang string, orderIndex int32) (QuestionSeed, error) {
	srcText := cell(row, idx, "srcText")
	if srcText == "" {
		return QuestionSeed{}, fmt.Errorf("id %s: empty srcText: %w", id, ErrInvalidCSVRow)
	}

	var segments []orderingSegmentJSON
	if err := json.Unmarshal([]byte(cell(row, idx, "segments")), &segments); err != nil {
		return QuestionSeed{}, fmt.Errorf("id %s: segments is not a JSON array of segments: %w: %w", id, ErrInvalidCSVRow, err)
	}
	if len(segments) == 0 {
		return QuestionSeed{}, fmt.Errorf("id %s: empty segments: %w", id, ErrInvalidCSVRow)
	}

	tags, err := csvRowTags(row, idx, "grammar:"+strings.ToLower(cell(row, idx, "pattern")))
	if err != nil {
		return QuestionSeed{}, fmt.Errorf("id %s: %w", id, err)
	}

	content, err := marshalContent(orderingContentJSON{
		Source:      textWithLangJSON{Text: srcText, Lang: sourceLang},
		Target:      orderingTargetJSON{Lang: targetLang, Segments: segments},
		Explanation: orderingExplanation(cell(row, idx, "patternName"), cell(row, idx, "patternPoint")),
	})
	if err != nil {
		return QuestionSeed{}, fmt.Errorf("id %s: %w", id, err)
	}

	return QuestionSeed{
		SeedKey:      id,
		QuestionType: questionTypeOrdering,
		Content:      content,
		Tags:         tags,
		OrderIndex:   orderIndex,
	}, nil
}

// orderingExplanation renders "**<patternName>** <patternPoint>", leaving out
// whichever part is empty.
func orderingExplanation(patternName, patternPoint string) string {
	var parts []string
	if patternName != "" {
		parts = append(parts, "**"+patternName+"**")
	}
	if patternPoint != "" {
		parts = append(parts, patternPoint)
	}
	return strings.Join(parts, " ")
}
