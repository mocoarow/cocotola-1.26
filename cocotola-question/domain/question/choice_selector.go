package question

import (
	"crypto/rand"
	"math/big"
)

// ChoiceSelector is a function type that selects choices for display.
// It can be swapped for alternative strategies in the future.
type ChoiceSelector func(choices []ChoiceJSON, displayCount int, shuffleChoices bool) []ChoiceJSON

// DefaultChoiceSelector selects all correct choices plus randomly selected
// incorrect ones up to displayCount.
func DefaultChoiceSelector(choices []ChoiceJSON, displayCount int, shuffleChoices bool) []ChoiceJSON {
	var correct, incorrect []ChoiceJSON
	for _, c := range choices {
		if c.IsCorrect {
			correct = append(correct, c)
		} else {
			incorrect = append(incorrect, c)
		}
	}

	cryptoShuffle(incorrect)

	needed := min(max(displayCount-len(correct), 0), len(incorrect))

	result := make([]ChoiceJSON, 0, displayCount)
	result = append(result, correct...)
	result = append(result, incorrect[:needed]...)

	if shuffleChoices {
		cryptoShuffle(result)
	}

	return result
}

func cryptoShuffle(s []ChoiceJSON) {
	for i := len(s) - 1; i > 0; i-- {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		j := int(n.Int64())
		s[i], s[j] = s[j], s[i]
	}
}
