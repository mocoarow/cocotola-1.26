package question

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
)

const (
	// MaxOrderingSegments is the maximum number of segments in ordering content.
	MaxOrderingSegments = 40
	// MaxOrderingTextLength is the maximum byte length of an ordering segment text.
	MaxOrderingTextLength = 500

	minOrderingChips = 2

	orderingSegmentChip  = "chip"
	orderingSegmentFixed = "fixed"

	orderingChipForbiddenChars  = "/[]"
	orderingFixedForbiddenChars = "[]"
)

// OrderingContent is the content of an ordering question.
type OrderingContent struct {
	Source      TextWithLang
	Target      OrderingTarget
	Explanation string
}

// OrderingTarget is the correct sentence split into segments.
type OrderingTarget struct {
	Lang     string
	Segments []OrderingSegment
}

// OrderingSegment is a chip to be ordered or a fixed text, by Type ("chip" or "fixed").
type OrderingSegment struct {
	Type string
	Text string
}

// Validate checks the ordering content rules.
func (c *OrderingContent) Validate() error {
	if err := validateTextWithLang(c.Source, "source"); err != nil {
		return fmt.Errorf("validate ordering source: %w", err)
	}
	if _, err := NewLang(c.Target.Lang); err != nil {
		return fmt.Errorf("invalid target language: %w", err)
	}
	if c.Source.Lang == c.Target.Lang {
		return fmt.Errorf("source and target language must differ: %w", domain.ErrInvalidArgument)
	}
	if err := validateOrderingSegments(c.Target.Segments); err != nil {
		return fmt.Errorf("validate ordering segments: %w", err)
	}
	if err := validateExplanation(c.Explanation); err != nil {
		return fmt.Errorf("validate ordering explanation: %w", err)
	}
	return nil
}

func validateOrderingSegments(segments []OrderingSegment) error {
	if len(segments) > MaxOrderingSegments {
		return fmt.Errorf("segments must not exceed %d: %w", MaxOrderingSegments, domain.ErrInvalidArgument)
	}
	chips := 0
	for i, s := range segments {
		if err := validateOrderingSegment(s); err != nil {
			return fmt.Errorf("segment[%d]: %w", i, err)
		}
		if s.Type == orderingSegmentChip {
			chips++
		}
	}
	if chips < minOrderingChips {
		return fmt.Errorf("at least %d chips are required: %w", minOrderingChips, domain.ErrInvalidArgument)
	}
	return nil
}

func validateOrderingSegment(s OrderingSegment) error {
	var forbidden string

	switch s.Type {
	case orderingSegmentChip:
		forbidden = orderingChipForbiddenChars
	case orderingSegmentFixed:
		forbidden = orderingFixedForbiddenChars
	default:
		return fmt.Errorf("type must be %q or %q: %w", orderingSegmentChip, orderingSegmentFixed, domain.ErrInvalidArgument)
	}
	if s.Text == "" {
		return fmt.Errorf("text is required: %w", domain.ErrInvalidArgument)
	}
	if strings.TrimSpace(s.Text) != s.Text {
		return fmt.Errorf("text must not have leading or trailing spaces: %w", domain.ErrInvalidArgument)
	}
	if len(s.Text) > MaxOrderingTextLength {
		return fmt.Errorf("text must not exceed %d bytes: %w", MaxOrderingTextLength, domain.ErrInvalidArgument)
	}
	if htmlTagPattern.MatchString(s.Text) {
		return fmt.Errorf("text must not contain HTML tags: %w", domain.ErrInvalidArgument)
	}
	if strings.ContainsAny(s.Text, forbidden) {
		return fmt.Errorf("%s text must not contain any of %q: %w", s.Type, forbidden, domain.ErrInvalidArgument)
	}
	return nil
}

// Regions returns the texts of each run of consecutive chips.
func (c *OrderingContent) Regions() [][]string {
	regions := make([][]string, 0, len(c.Target.Segments))
	var current []string
	for _, s := range c.Target.Segments {
		if s.Type == orderingSegmentChip {
			current = append(current, s.Text)
			continue
		}
		if len(current) > 0 {
			regions = append(regions, current)
			current = nil
		}
	}
	if len(current) > 0 {
		regions = append(regions, current)
	}
	return regions
}

// EvaluateAnswer reports whether the answer has the same regions with the same texts in the same order.
// An answer of a different shape is incorrect, not an error.
func (c *OrderingContent) EvaluateAnswer(regions [][]string) (bool, error) {
	want := c.Regions()
	if len(want) == 0 {
		return false, fmt.Errorf("ordering content has no chip: %w", domain.ErrInvalidArgument)
	}
	if len(regions) != len(want) {
		return false, nil
	}
	for i := range want {
		if !slices.Equal(regions[i], want[i]) {
			return false, nil
		}
	}
	return true, nil
}
