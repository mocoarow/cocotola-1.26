//go:build small

package seed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

func Test_DefaultSeeds_shouldReturnNonEmpty_andHaveUniqueSeedKeys(t *testing.T) {
	t.Parallel()

	// when
	seeds, err := seed.DefaultSeeds()

	// then
	require.NoError(t, err)
	require.NotEmpty(t, seeds, "default seeds must not be empty")

	keys := make(map[string]bool, len(seeds))
	for _, s := range seeds {
		assert.NotEmptyf(t, s.SeedKey, "seedKey must be non-empty for %q", s.Title)
		assert.NotEmptyf(t, s.Title, "title must be non-empty for %q", s.SeedKey)
		assert.Falsef(t, keys[s.SeedKey], "duplicate seedKey %q", s.SeedKey)
		keys[s.SeedKey] = true
	}
}

func Test_validateSeeds_shouldReturnNil_whenLanguageIsISO6391(t *testing.T) {
	t.Parallel()

	// given
	workbooks := []seed.PublicWorkbookSeed{{SeedKey: "wb-1", Language: "ja"}}

	// when
	err := seed.ValidateSeeds(workbooks)

	// then
	require.NoError(t, err)
}

func Test_validateSeeds_shouldReturnErrInvalidSeed_whenLanguageIsNotISO6391(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		language string
	}{
		{name: "empty", language: ""},
		{name: "uppercase", language: "EN"},
		{name: "english name", language: "Japanese"},
		{name: "region subtag", language: "ja-JP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			workbooks := []seed.PublicWorkbookSeed{{SeedKey: "wb-1", Language: tt.language}}

			// when
			err := seed.ValidateSeeds(workbooks)

			// then
			require.ErrorIs(t, err, seed.ErrInvalidSeed)
		})
	}
}
