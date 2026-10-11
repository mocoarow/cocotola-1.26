package contentjson

import (
	"encoding/json"
	"fmt"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

type textWithLang struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

type wordFill struct {
	Source             textWithLang `json:"source"`
	Target             textWithLang `json:"target"`
	Explanation        string       `json:"explanation"`
	Explanation1       string       `json:"explanation1"`
	Explanation2       string       `json:"explanation2"`
	AllowPartialCredit bool         `json:"allowPartialCredit"`
}

type choice struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}

type multipleChoice struct {
	QuestionText       string   `json:"questionText"`
	Explanation        string   `json:"explanation"`
	Choices            []choice `json:"choices"`
	DisplayCount       int      `json:"displayCount"`
	ShowCorrectCount   bool     `json:"showCorrectCount"`
	ShuffleChoices     bool     `json:"shuffleChoices"`
	AllowPartialCredit bool     `json:"allowPartialCredit"`
}

// Parser implements question.ContentParser for the JSON content format.
type Parser struct{}

// ParseWordFill parses word_fill content JSON.
func (Parser) ParseWordFill(raw string) (*question.WordFillContent, error) {
	var c wordFill
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, fmt.Errorf("invalid word_fill content JSON: %w", domain.ErrInvalidArgument)
	}
	return &question.WordFillContent{
		Source:             question.TextWithLang{Text: c.Source.Text, Lang: c.Source.Lang},
		Target:             question.TextWithLang{Text: c.Target.Text, Lang: c.Target.Lang},
		Explanation:        c.Explanation,
		Explanation1:       c.Explanation1,
		Explanation2:       c.Explanation2,
		AllowPartialCredit: c.AllowPartialCredit,
	}, nil
}

// ParseMultipleChoice parses multiple_choice content JSON.
func (Parser) ParseMultipleChoice(raw string) (*question.MultipleChoiceContent, error) {
	var c multipleChoice
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return nil, fmt.Errorf("invalid multiple_choice content JSON: %w", domain.ErrInvalidArgument)
	}
	choices := make([]question.Choice, len(c.Choices))
	for i, ch := range c.Choices {
		choices[i] = question.Choice{ID: ch.ID, Text: ch.Text, IsCorrect: ch.IsCorrect}
	}
	return &question.MultipleChoiceContent{
		QuestionText:       c.QuestionText,
		Explanation:        c.Explanation,
		Choices:            choices,
		DisplayCount:       c.DisplayCount,
		ShowCorrectCount:   c.ShowCorrectCount,
		ShuffleChoices:     c.ShuffleChoices,
		AllowPartialCredit: c.AllowPartialCredit,
	}, nil
}
