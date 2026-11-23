package seed

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"
)

const (
	// formatTatoebaWordFill is the CSV format whose rows hold a Japanese source
	// sentence plus an English sentence with "___" blanks and the comma-joined
	// answers. Each row becomes one word_fill question.
	formatTatoebaWordFill = "tatoeba-wordfill"

	// questionTypeWordFill mirrors cocotola-question's word_fill type value.
	// It is duplicated here so the seed package does not depend on the question
	// domain just for this constant.
	questionTypeWordFill = "word_fill"

	// blankPlaceholder is the token used in the CSV blankText column for each
	// missing word; it is replaced by the {{answer}} form word_fill expects.
	blankPlaceholder = "___"

	// utf8BOM is stripped from the first CSV header cell when present.
	utf8BOM = "\ufeff"

	// csvHeaderRows is the number of leading rows occupied by the header.
	csvHeaderRows = 1

	// Tatoeba attribution building blocks. Sentences imported from Tatoeba are
	// licensed CC BY 2.0 FR and must credit the sentence id and author.
	tatoebaSentenceURL = "https://tatoeba.org/en/sentences/show/"
	tatoebaProfileURL  = "https://tatoeba.org/en/user/profile/"
	tatoebaLicenseName = "CC BY 2.0 FR"
	tatoebaLicenseURL  = "https://creativecommons.org/licenses/by/2.0/fr/"
)

// CSV converter errors. They are sentinels so callers (and tests) can match
// them with errors.Is regardless of the surrounding context message.
var (
	// ErrUnsupportedCSVFormat is returned when a manifest entry names a format
	// the converter does not know how to handle.
	ErrUnsupportedCSVFormat = errors.New("unsupported csv format")
	// ErrMissingCSVColumn is returned when the CSV header lacks a required column.
	ErrMissingCSVColumn = errors.New("missing csv column")
	// ErrInvalidCSVRow marks a data row that cannot be converted, or a CSV
	// without any data row.
	ErrInvalidCSVRow = errors.New("invalid csv row")
)

// tagValuePattern is the part of cocotola-question's tag pattern on each side
// of the colon. Row ids and derived tag values must match it because they end
// up in question tags.
var tagValuePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// wordFillContentJSON mirrors the word_fill JSON of cocotola-question/contentjson
// without importing it, so the seed package has no cocotola-question dependency.
type wordFillContentJSON struct {
	Source             textWithLangJSON `json:"source"`
	Target             textWithLangJSON `json:"target"`
	Explanation1       string           `json:"explanation1"`
	Explanation2       string           `json:"explanation2"`
	AllowPartialCredit bool             `json:"allowPartialCredit"`
}

type textWithLangJSON struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

// csvConversion is the result of converting one CSV: the questions from valid
// rows and the rows skipped as invalid.
type csvConversion struct {
	questions []QuestionSeed
	skipped   []skippedCSVRow
}

// skippedCSVRow identifies an invalid row by its position, on the same basis
// as QuestionSeed.OrderIndex.
type skippedCSVRow struct {
	index  int32
	reason error
}

// convertCSV converts raw CSV bytes into question seeds according to format.
func convertCSV(format, sourceLang, targetLang string, data []byte) (csvConversion, error) {
	switch format {
	case formatTatoebaWordFill:
		return convertWordFillCSV(sourceLang, targetLang, data)
	case formatGeneratedOrdering:
		return convertOrderingCSV(sourceLang, targetLang, data)
	default:
		return csvConversion{}, fmt.Errorf("format %q: %w", format, ErrUnsupportedCSVFormat)
	}
}

// csvRowConverter converts one data row, whose id is already checked, into a
// QuestionSeed or returns an ErrInvalidCSVRow-wrapped error.
type csvRowConverter func(row []string, idx map[string]int, id string, orderIndex int32) (QuestionSeed, error)

// convertCSVRows maps each valid data row to a QuestionSeed and skips invalid
// rows. The CSV `id` column becomes the question seedKey (stable across runs so
// appended rows are detected as new), and OrderIndex is the row's position, so
// it stays the same as the CSV grows.
func convertCSVRows(data []byte, requiredColumns []string, convertRow csvRowConverter) (csvConversion, error) {
	header, rows, err := readCSVRows(data, requiredColumns)
	if err != nil {
		return csvConversion{}, err
	}
	idx := indexColumns(header)

	questions := make([]QuestionSeed, 0, len(rows))
	var skipped []skippedCSVRow
	seen := make(map[string]bool, len(rows))
	for i, row := range rows {
		index := int32(i + 1)
		id, err := checkRowID(row, len(header), idx, seen)
		if err != nil {
			skipped = append(skipped, skippedCSVRow{index: index, reason: err})
			continue
		}
		q, err := convertRow(row, idx, id, index)
		if err != nil {
			skipped = append(skipped, skippedCSVRow{index: index, reason: err})
			continue
		}
		seen[id] = true
		questions = append(questions, q)
	}
	return csvConversion{questions: questions, skipped: skipped}, nil
}

// readCSVRows parses data and returns the header and the data rows. Rows may
// have any number of fields so that a row with the wrong count can be skipped
// instead of failing the whole CSV.
func readCSVRows(data []byte, requiredColumns []string) ([]string, [][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("read csv: %w", err)
	}
	if len(records) <= csvHeaderRows {
		return nil, nil, fmt.Errorf("csv needs a header and at least one data row: %w", ErrInvalidCSVRow)
	}

	if err := requireColumns(indexColumns(records[0]), requiredColumns...); err != nil {
		return nil, nil, err
	}
	return records[0], records[csvHeaderRows:], nil
}

// checkRowID returns the row's id when the row has as many fields as the
// header, the id matches the tag pattern, and no accepted row (seen) has it.
func checkRowID(row []string, headerLen int, idx map[string]int, seen map[string]bool) (string, error) {
	if len(row) != headerLen {
		return "", fmt.Errorf("%d fields, header has %d: %w", len(row), headerLen, ErrInvalidCSVRow)
	}

	id := cell(row, idx, "id")
	switch {
	case !tagValuePattern.MatchString(id):
		return "", fmt.Errorf("id is empty or outside the tag pattern: %w", ErrInvalidCSVRow)
	case seen[id]:
		return "", fmt.Errorf("duplicate id %s: %w", id, ErrInvalidCSVRow)
	}
	return id, nil
}

// convertWordFillCSV maps each valid data row to a word_fill QuestionSeed.
func convertWordFillCSV(sourceLang, targetLang string, data []byte) (csvConversion, error) {
	return convertCSVRows(data, []string{"id", "srcText", "blankText", "blankAnswers"},
		func(row []string, idx map[string]int, id string, orderIndex int32) (QuestionSeed, error) {
			return convertWordFillRow(row, idx, id, sourceLang, targetLang, orderIndex)
		})
}

// convertWordFillRow converts a single CSV data row into a word_fill QuestionSeed.
func convertWordFillRow(row []string, idx map[string]int, id, sourceLang, targetLang string, orderIndex int32) (QuestionSeed, error) {
	if cell(row, idx, "srcText") == "" {
		return QuestionSeed{}, fmt.Errorf("id %s: empty srcText: %w", id, ErrInvalidCSVRow)
	}

	tags, err := csvRowTags(row, idx)
	if err != nil {
		return QuestionSeed{}, fmt.Errorf("id %s: %w", id, err)
	}

	answers := splitCommaList(cell(row, idx, "blankAnswers"))
	target, err := buildWordFillTarget(cell(row, idx, "blankText"), answers)
	if err != nil {
		return QuestionSeed{}, fmt.Errorf("id %s: %w", id, err)
	}

	// In this dataset the src* columns describe the target-language (English)
	// sentence and dst* columns the source-language (Japanese) sentence, so
	// columns are paired with the opposite language.
	attribution := tatoebaAttribution([]sentenceCitation{
		{lang: sourceLang, number: cell(row, idx, "dstSentenceNumber"), author: cell(row, idx, "dstAuthor")},
		{lang: targetLang, number: cell(row, idx, "srcSentenceNumber"), author: cell(row, idx, "srcAuthor")},
	})

	content, err := marshalContent(wordFillContentJSON{
		Source:             textWithLangJSON{Text: cell(row, idx, "srcText"), Lang: sourceLang},
		Target:             textWithLangJSON{Text: target, Lang: targetLang},
		Explanation1:       attribution,
		Explanation2:       attribution,
		AllowPartialCredit: len(answers) > 1,
	})
	if err != nil {
		return QuestionSeed{}, fmt.Errorf("id %s: %w", id, err)
	}

	return QuestionSeed{
		SeedKey:      id,
		QuestionType: questionTypeWordFill,
		Content:      content,
		Tags:         tags,
		OrderIndex:   orderIndex,
	}, nil
}

// buildWordFillTarget replaces each "___" in blankText with the matching
// {{answer}}. The number of blanks must equal the number of answers.
func buildWordFillTarget(blankText string, answers []string) (string, error) {
	blanks := strings.Count(blankText, blankPlaceholder)
	if blanks == 0 {
		return "", fmt.Errorf("blankText has no %q placeholder: %w", blankPlaceholder, ErrInvalidCSVRow)
	}
	if blanks != len(answers) {
		return "", fmt.Errorf("blank count %d does not match answer count %d: %w", blanks, len(answers), ErrInvalidCSVRow)
	}

	result := blankText
	for _, answer := range answers {
		result = strings.Replace(result, blankPlaceholder, "{{"+answer+"}}", 1)
	}
	return result, nil
}

func marshalContent(content any) (string, error) {
	encoded, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("marshal content: %w", err)
	}
	return string(encoded), nil
}

// sentenceCitation is one Tatoeba sentence's attribution: its language label,
// sentence id (number), and author (may be empty).
type sentenceCitation struct {
	lang   string
	number string
	author string
}

// tatoebaAttribution renders Markdown attribution lines (one per citation with
// a non-empty number), joined by a blank line. Citations without a sentence id
// are skipped. The author segment is omitted when the author is unknown.
func tatoebaAttribution(citations []sentenceCitation) string {
	lines := make([]string, 0, len(citations))
	for _, c := range citations {
		if c.number == "" {
			continue
		}
		line := fmt.Sprintf("Sentence source(%s): Tatoeba [#%s](%s%s)",
			c.lang, c.number, tatoebaSentenceURL, url.PathEscape(c.number))
		if c.author != "" {
			line += fmt.Sprintf(" / Author: [%s](%s%s)",
				c.author, tatoebaProfileURL, url.PathEscape(c.author))
		}
		line += fmt.Sprintf(" / Licensed under [%s](%s)", tatoebaLicenseName, tatoebaLicenseURL)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n\n")
}

// csvRowTags derives extra question tags from the CSV: one level:<value> tag
// per entry in the tags column, then extra, then a source:<value> tag from
// sourceType. The seeder prepends its own seed identity tag separately.
func csvRowTags(row []string, idx map[string]int, extra ...string) ([]string, error) {
	var tags []string
	for _, level := range splitCommaList(cell(row, idx, "tags")) {
		if tag := "level:" + strings.ToLower(level); !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}
	tags = append(tags, extra...)
	if source := cell(row, idx, "sourceType"); source != "" {
		tags = append(tags, "source:"+strings.ToLower(source))
	}
	for _, tag := range tags {
		if _, value, _ := strings.Cut(tag, ":"); !tagValuePattern.MatchString(value) {
			return nil, fmt.Errorf("tag value outside the tag pattern: %w", ErrInvalidCSVRow)
		}
	}
	return tags, nil
}

// indexColumns maps each header name to its column position, tolerating a
// UTF-8 BOM on the first column and surrounding whitespace.
func indexColumns(header []string) map[string]int {
	result := make(map[string]int, len(header))
	for i, name := range header {
		result[strings.TrimSpace(strings.TrimPrefix(name, utf8BOM))] = i
	}
	return result
}

func requireColumns(idx map[string]int, names ...string) error {
	for _, name := range names {
		if _, ok := idx[name]; !ok {
			return fmt.Errorf("column %q: %w", name, ErrMissingCSVColumn)
		}
	}
	return nil
}

// cell returns the trimmed value of the named column for a row, or "" when the
// column is absent or the row is short.
func cell(row []string, idx map[string]int, name string) string {
	i, ok := idx[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

// splitCommaList splits a comma-separated value, trimming and dropping empties.
func splitCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
