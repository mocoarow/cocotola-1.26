package question

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	maxQuestionTextLength = 5000
	maxExplanationLength  = 5000
	maxChoiceTextLength   = 500
	maxChoices            = 40
	maxDisplayCount       = 20
	minChoices            = 1
	minDisplayCount       = 1
)

var (
	blankPattern   = regexp.MustCompile(`\{\{([^}]+)\}\}`)
	htmlTagPattern = regexp.MustCompile(`<[^>]+>`)
)

// --- Shared ---

// ScoringModeJSON represents the scoring mode in content JSON.
type ScoringModeJSON struct {
	AllowPartialCredit bool `json:"allowPartialCredit"`
}

// --- word_fill ---

// WordFillContent represents the JSON content for a word_fill question.
type WordFillContent struct {
	Source             TextWithLang `json:"source"`
	Target             TextWithLang `json:"target"`
	Explanation        string       `json:"explanation"`
	AllowPartialCredit bool         `json:"allowPartialCredit"`
}

// TextWithLang represents a text string with its language code.
type TextWithLang struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

// --- multiple_choice ---

// MultipleChoiceContent represents the JSON content for a multiple_choice question.
type MultipleChoiceContent struct {
	QuestionText       string       `json:"questionText"`
	Explanation        string       `json:"explanation"`
	Choices            []ChoiceJSON `json:"choices"`
	DisplayCount       int          `json:"displayCount"`
	ShowCorrectCount   bool         `json:"showCorrectCount"`
	ShuffleChoices     bool         `json:"shuffleChoices"`
	AllowPartialCredit bool         `json:"allowPartialCredit"`
}

// ChoiceJSON represents a single choice in a multiple_choice question.
type ChoiceJSON struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}

// ValidateContent validates the content JSON based on question type.
func ValidateContent(questionType Type, content string) error {
	switch questionType.Value() {
	case questionTypeWordFill:
		return validateWordFillContent(content)
	case questionTypeMultipleChoice:
		return validateMultipleChoiceContent(content)
	default:
		return fmt.Errorf("unknown question type for content validation: %w", domain.ErrInvalidArgument)
	}
}

func validateWordFillContent(content string) error {
	var c WordFillContent
	if err := json.Unmarshal([]byte(content), &c); err != nil {
		return fmt.Errorf("invalid word_fill content JSON: %w", domain.ErrInvalidArgument)
	}

	if err := validateTextWithLang(c.Source, "source"); err != nil {
		return err
	}
	if err := validateTextWithLang(c.Target, "target"); err != nil {
		return err
	}

	if c.Source.Lang == c.Target.Lang {
		return fmt.Errorf("source and target language must differ: %w", domain.ErrInvalidArgument)
	}

	blanks := blankPattern.FindAllStringSubmatch(c.Target.Text, -1)
	if len(blanks) == 0 {
		return fmt.Errorf("target text must contain at least one {{word}} blank: %w", domain.ErrInvalidArgument)
	}
	for _, blank := range blanks {
		if strings.TrimSpace(blank[1]) == "" {
			return fmt.Errorf("blank word must not be empty: %w", domain.ErrInvalidArgument)
		}
	}

	if err := validateExplanation(c.Explanation); err != nil {
		return err
	}

	return nil
}

func validateExplanation(explanation string) error {
	if explanation == "" {
		return nil
	}
	if len(explanation) > maxExplanationLength {
		return fmt.Errorf("explanation must not exceed %d characters: %w", maxExplanationLength, domain.ErrInvalidArgument)
	}
	if htmlTagPattern.MatchString(explanation) {
		return fmt.Errorf("explanation must not contain HTML tags: %w", domain.ErrInvalidArgument)
	}
	return nil
}

func validateTextWithLang(twl TextWithLang, fieldName string) error {
	if twl.Text == "" {
		return fmt.Errorf("%s text is required: %w", fieldName, domain.ErrInvalidArgument)
	}
	if len(twl.Text) > maxQuestionTextLength {
		return fmt.Errorf("%s text must not exceed %d characters: %w", fieldName, maxQuestionTextLength, domain.ErrInvalidArgument)
	}
	if htmlTagPattern.MatchString(twl.Text) {
		return fmt.Errorf("%s text must not contain HTML tags: %w", fieldName, domain.ErrInvalidArgument)
	}
	if _, err := NewLang(twl.Lang); err != nil {
		return fmt.Errorf("invalid %s language: %w", fieldName, err)
	}
	return nil
}

func validateMultipleChoiceContent(content string) error {
	var c MultipleChoiceContent
	if err := json.Unmarshal([]byte(content), &c); err != nil {
		return fmt.Errorf("invalid multiple_choice content JSON: %w", domain.ErrInvalidArgument)
	}

	if err := validateQuestionText(c.QuestionText); err != nil {
		return err
	}
	if err := validateExplanation(c.Explanation); err != nil {
		return err
	}
	if err := validateChoices(c.Choices); err != nil {
		return err
	}
	if err := validateDisplayCount(c.DisplayCount, c.Choices); err != nil {
		return err
	}

	return nil
}

func validateQuestionText(text string) error {
	if text == "" {
		return fmt.Errorf("questionText is required: %w", domain.ErrInvalidArgument)
	}
	if len(text) > maxQuestionTextLength {
		return fmt.Errorf("questionText must not exceed %d characters: %w", maxQuestionTextLength, domain.ErrInvalidArgument)
	}
	if htmlTagPattern.MatchString(text) {
		return fmt.Errorf("questionText must not contain HTML tags: %w", domain.ErrInvalidArgument)
	}
	return nil
}

func validateChoices(choices []ChoiceJSON) error {
	if len(choices) < minChoices {
		return fmt.Errorf("at least %d choice is required: %w", minChoices, domain.ErrInvalidArgument)
	}
	if len(choices) > maxChoices {
		return fmt.Errorf("choices must not exceed %d: %w", maxChoices, domain.ErrInvalidArgument)
	}

	idSet := make(map[string]bool, len(choices))
	for i, choice := range choices {
		if choice.ID == "" {
			return fmt.Errorf("choice[%d] id is required: %w", i, domain.ErrInvalidArgument)
		}
		if idSet[choice.ID] {
			return fmt.Errorf("choice id %q is duplicated: %w", choice.ID, domain.ErrInvalidArgument)
		}
		idSet[choice.ID] = true

		if choice.Text == "" {
			return fmt.Errorf("choice[%d] text is required: %w", i, domain.ErrInvalidArgument)
		}
		if len(choice.Text) > maxChoiceTextLength {
			return fmt.Errorf("choice[%d] text must not exceed %d characters: %w", i, maxChoiceTextLength, domain.ErrInvalidArgument)
		}
	}

	correctCount := countCorrectChoices(choices)
	if correctCount == 0 {
		return fmt.Errorf("at least one choice must be correct: %w", domain.ErrInvalidArgument)
	}

	return nil
}

func validateDisplayCount(displayCount int, choices []ChoiceJSON) error {
	if displayCount < minDisplayCount || displayCount > maxDisplayCount {
		return fmt.Errorf("displayCount must be between %d and %d: %w", minDisplayCount, maxDisplayCount, domain.ErrInvalidArgument)
	}
	if displayCount > len(choices) {
		return fmt.Errorf("displayCount must not exceed number of choices: %w", domain.ErrInvalidArgument)
	}
	correctCount := countCorrectChoices(choices)
	if displayCount < correctCount {
		return fmt.Errorf("displayCount must be >= number of correct choices (%d): %w", correctCount, domain.ErrInvalidArgument)
	}
	return nil
}

func countCorrectChoices(choices []ChoiceJSON) int {
	count := 0
	for _, c := range choices {
		if c.IsCorrect {
			count++
		}
	}
	return count
}
