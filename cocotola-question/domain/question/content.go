package question

import (
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

// ContentParser parses the raw content of a question. Parsing does not validate.
type ContentParser interface {
	ParseWordFill(raw string) (*WordFillContent, error)
	ParseMultipleChoice(raw string) (*MultipleChoiceContent, error)
}

// --- word_fill ---

// WordFillContent is the content of a word_fill question.
// Explanation is legacy; Explanation1 is shown with the question and Explanation2 with the answer
// (both carry attribution/licensing for imported sentences).
type WordFillContent struct {
	Source             TextWithLang
	Target             TextWithLang
	Explanation        string
	Explanation1       string
	Explanation2       string
	AllowPartialCredit bool
}

// TextWithLang represents a text string with its language code.
type TextWithLang struct {
	Text string
	Lang string
}

// --- multiple_choice ---

// MultipleChoiceContent is the content of a multiple_choice question.
type MultipleChoiceContent struct {
	QuestionText       string
	Explanation        string
	Choices            []Choice
	DisplayCount       int
	ShowCorrectCount   bool
	ShuffleChoices     bool
	AllowPartialCredit bool
}

// Choice represents a single choice in a multiple_choice question.
type Choice struct {
	ID        string
	Text      string
	IsCorrect bool
}

// EvaluateAnswer reports whether the given choice IDs exactly match the set of
// IsCorrect choices. Strict equality only — AllowPartialCredit is intentionally
// not consulted yet (a follow-up will add partial-credit scoring).
// Returns ErrInvalidArgument wrapped if any selected ID is not a known choice.
func (c *MultipleChoiceContent) EvaluateAnswer(selectedChoiceIDs []string) (bool, error) {
	valid := make(map[string]bool, len(c.Choices))
	correct := make(map[string]bool, len(c.Choices))
	for _, ch := range c.Choices {
		valid[ch.ID] = true
		if ch.IsCorrect {
			correct[ch.ID] = true
		}
	}
	selected := make(map[string]bool, len(selectedChoiceIDs))
	for _, id := range selectedChoiceIDs {
		if !valid[id] {
			return false, fmt.Errorf("unknown choice id %q: %w", id, domain.ErrInvalidArgument)
		}
		selected[id] = true
	}
	if len(selected) != len(correct) {
		return false, nil
	}
	for id := range correct {
		if !selected[id] {
			return false, nil
		}
	}
	return true, nil
}

// Validate checks the word_fill content rules.
func (c *WordFillContent) Validate() error {
	if err := validateTextWithLang(c.Source, "source"); err != nil {
		return fmt.Errorf("validate word_fill source: %w", err)
	}
	if err := validateTextWithLang(c.Target, "target"); err != nil {
		return fmt.Errorf("validate word_fill target: %w", err)
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
		return fmt.Errorf("validate word_fill explanation: %w", err)
	}
	if err := validateExplanation(c.Explanation1); err != nil {
		return fmt.Errorf("validate word_fill explanation1: %w", err)
	}
	if err := validateExplanation(c.Explanation2); err != nil {
		return fmt.Errorf("validate word_fill explanation2: %w", err)
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

// Validate checks the multiple_choice content rules.
func (c *MultipleChoiceContent) Validate() error {
	if err := validateQuestionText(c.QuestionText); err != nil {
		return fmt.Errorf("validate multiple_choice question text: %w", err)
	}
	if err := validateExplanation(c.Explanation); err != nil {
		return fmt.Errorf("validate multiple_choice explanation: %w", err)
	}
	if err := validateChoices(c.Choices); err != nil {
		return fmt.Errorf("validate multiple_choice choices: %w", err)
	}
	if err := validateDisplayCount(c.DisplayCount, c.Choices); err != nil {
		return fmt.Errorf("validate multiple_choice display count: %w", err)
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

func validateChoices(choices []Choice) error {
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

func validateDisplayCount(displayCount int, choices []Choice) error {
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

func countCorrectChoices(choices []Choice) int {
	count := 0
	for _, c := range choices {
		if c.IsCorrect {
			count++
		}
	}
	return count
}
