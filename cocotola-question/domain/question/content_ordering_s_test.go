package question_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain"
	"github.com/mocoarow/cocotola-1.26/cocotola-question/domain/question"
)

type orderingCaseSegment struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type orderingCase struct {
	Name     string                `json:"name"`
	Segments []orderingCaseSegment `json:"segments"`
	Regions  [][]string            `json:"regions"`
	Answer   [][]string            `json:"answer"`
	Correct  bool                  `json:"correct"`
}

func loadOrderingCases(t *testing.T) []orderingCase {
	t.Helper()
	raw, err := os.ReadFile("testdata/ordering_cases.json")
	require.NoError(t, err)
	var cases []orderingCase
	require.NoError(t, json.Unmarshal(raw, &cases))
	require.NotEmpty(t, cases)
	return cases
}

func orderingContentFromCase(c orderingCase) *question.OrderingContent {
	segments := make([]question.OrderingSegment, len(c.Segments))
	for i, s := range c.Segments {
		segments[i] = question.OrderingSegment{Type: s.Type, Text: s.Text}
	}
	return &question.OrderingContent{
		Source: question.TextWithLang{Text: "原文", Lang: "ja"},
		Target: question.OrderingTarget{Lang: "en", Segments: segments},
	}
}

func chip(text string) question.OrderingSegment {
	return question.OrderingSegment{Type: "chip", Text: text}
}

func fixed(text string) question.OrderingSegment {
	return question.OrderingSegment{Type: "fixed", Text: text}
}

func validOrderingContent() *question.OrderingContent {
	return &question.OrderingContent{
		Source: question.TextWithLang{Text: "私は学生です。あなたは先生です。", Lang: "ja"},
		Target: question.OrderingTarget{
			Lang: "en",
			Segments: []question.OrderingSegment{
				chip("I"), chip("am"), chip("a student"), fixed("."),
				chip("you"), chip("are"), chip("a teacher"), fixed("."),
			},
		},
		Explanation: "be 動詞は主語に合わせます。",
	}
}

func withSegments(segments ...question.OrderingSegment) *question.OrderingContent {
	base := validOrderingContent()
	return &question.OrderingContent{
		Source:      base.Source,
		Target:      question.OrderingTarget{Lang: base.Target.Lang, Segments: segments},
		Explanation: base.Explanation,
	}
}

func repeatChips(n int) []question.OrderingSegment {
	segments := make([]question.OrderingSegment, n)
	for i := range segments {
		segments[i] = chip("w")
	}
	return segments
}

func Test_OrderingContent_Validate_shouldSucceed_whenContentIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content *question.OrderingContent
	}{
		{name: "spec example", content: validOrderingContent()},
		{name: "max segments", content: withSegments(repeatChips(question.MaxOrderingSegments)...)},
		{name: "max text bytes", content: withSegments(chip(strings.Repeat("a", question.MaxOrderingTextLength)), chip("b"))},
		{name: "fixed contains slash", content: withSegments(chip("a"), fixed("and/or"), chip("b"))},
		{name: "chips split into one-chip regions", content: withSegments(chip("a"), fixed(","), chip("b"))},
		{name: "no explanation", content: &question.OrderingContent{
			Source: question.TextWithLang{Text: "原文", Lang: "ja"},
			Target: question.OrderingTarget{Lang: "en", Segments: []question.OrderingSegment{chip("a"), chip("b")}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := tt.content.Validate()

			// then
			require.NoError(t, err)
		})
	}
}

func Test_OrderingContent_Validate_shouldReturnErrInvalidArgument_whenContentIsInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content *question.OrderingContent
	}{
		{name: "no segments", content: withSegments()},
		{name: "too many segments", content: withSegments(repeatChips(question.MaxOrderingSegments + 1)...)},
		{name: "unknown segment type", content: withSegments(chip("a"), chip("b"), question.OrderingSegment{Type: "blank", Text: "c"})},
		{name: "empty text", content: withSegments(chip("a"), chip(""))},
		{name: "leading space", content: withSegments(chip("a"), chip(" b"))},
		{name: "trailing space", content: withSegments(chip("a"), chip("b"), fixed(". "))},
		{name: "text too long", content: withSegments(chip(strings.Repeat("a", question.MaxOrderingTextLength+1)), chip("b"))},
		{name: "html tag", content: withSegments(chip("a"), chip("<b>"))},
		{name: "single chip", content: withSegments(chip("a"), fixed("."))},
		{name: "chip contains slash", content: withSegments(chip("a/b"), chip("c"))},
		{name: "chip contains open bracket", content: withSegments(chip("a["), chip("c"))},
		{name: "chip contains close bracket", content: withSegments(chip("a]"), chip("c"))},
		{name: "fixed contains open bracket", content: withSegments(chip("a"), fixed("["), chip("c"))},
		{name: "fixed contains close bracket", content: withSegments(chip("a"), fixed("]"), chip("c"))},
		{name: "empty source text", content: &question.OrderingContent{
			Source: question.TextWithLang{Text: "", Lang: "ja"},
			Target: question.OrderingTarget{Lang: "en", Segments: []question.OrderingSegment{chip("a"), chip("b")}},
		}},
		{name: "invalid source lang", content: &question.OrderingContent{
			Source: question.TextWithLang{Text: "原文", Lang: "xx"},
			Target: question.OrderingTarget{Lang: "en", Segments: []question.OrderingSegment{chip("a"), chip("b")}},
		}},
		{name: "invalid target lang", content: &question.OrderingContent{
			Source: question.TextWithLang{Text: "原文", Lang: "ja"},
			Target: question.OrderingTarget{Lang: "xx", Segments: []question.OrderingSegment{chip("a"), chip("b")}},
		}},
		{name: "same source and target lang", content: &question.OrderingContent{
			Source: question.TextWithLang{Text: "原文", Lang: "en"},
			Target: question.OrderingTarget{Lang: "en", Segments: []question.OrderingSegment{chip("a"), chip("b")}},
		}},
		{name: "explanation has html", content: &question.OrderingContent{
			Source:      question.TextWithLang{Text: "原文", Lang: "ja"},
			Target:      question.OrderingTarget{Lang: "en", Segments: []question.OrderingSegment{chip("a"), chip("b")}},
			Explanation: "<i>x</i>",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// when
			err := tt.content.Validate()

			// then
			require.ErrorIs(t, err, domain.ErrInvalidArgument)
		})
	}
}

func Test_OrderingContent_Regions_shouldReturnConsecutiveChipRuns_whenFixtureIsGiven(t *testing.T) {
	t.Parallel()

	for _, tt := range loadOrderingCases(t) {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			// given
			c := orderingContentFromCase(tt)

			// when
			got := c.Regions()

			// then
			assert.Equal(t, tt.Regions, got)
		})
	}
}

func Test_OrderingContent_EvaluateAnswer_shouldReturnExpected_whenFixtureIsGiven(t *testing.T) {
	t.Parallel()

	for _, tt := range loadOrderingCases(t) {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			// given
			c := orderingContentFromCase(tt)

			// when
			got, err := c.EvaluateAnswer(tt.Answer)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.Correct, got)
		})
	}
}

func Test_OrderingContent_EvaluateAnswer_shouldReturnErrInvalidArgument_whenContentHasNoChip(t *testing.T) {
	t.Parallel()

	// given
	c := withSegments(fixed("."))

	// when
	_, err := c.EvaluateAnswer([][]string{{"a"}})

	// then
	require.ErrorIs(t, err, domain.ErrInvalidArgument)
}
