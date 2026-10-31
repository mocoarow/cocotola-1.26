//go:build small

package seed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

func Test_DefaultCSVManifest_shouldParseEmbeddedFile_withUniqueAndCompleteEntries(t *testing.T) {
	t.Parallel()

	// when
	manifest, err := seed.DefaultCSVManifest()

	// then
	require.NoError(t, err)

	keys := make(map[string]bool, len(manifest.Workbooks))
	for _, e := range manifest.Workbooks {
		assert.NotEmptyf(t, e.SeedKey, "seedKey must be non-empty for %q", e.Title)
		assert.NotEmptyf(t, e.Title, "title must be non-empty for %q", e.SeedKey)
		assert.NotEmptyf(t, e.Format, "format must be non-empty for %q", e.SeedKey)
		assert.NotEmptyf(t, e.GCSObject, "gcsObject must be non-empty for %q", e.SeedKey)
		assert.NotEqualf(t, e.SourceLang, e.TargetLang, "sourceLang and targetLang must differ for %q", e.SeedKey)
		assert.Falsef(t, keys[e.SeedKey], "duplicate seedKey %q", e.SeedKey)
		keys[e.SeedKey] = true
	}
}

func Test_CSVWorkbookManifest_validate_shouldReturnNil_whenLanguagesAreISO6391(t *testing.T) {
	t.Parallel()

	// given
	manifest := wordFillManifest()

	// when
	err := seed.ValidateCSVManifest(manifest)

	// then
	require.NoError(t, err)
}

func Test_CSVWorkbookManifest_validate_shouldReturnErrInvalidManifest_whenLanguageCodeIsNotISO6391(t *testing.T) {
	t.Parallel()

	withLanguage := func(v string) seed.CSVWorkbookEntry {
		e := wordFillManifest().Workbooks[0]
		e.Language = v
		return e
	}
	withSourceLang := func(v string) seed.CSVWorkbookEntry {
		e := wordFillManifest().Workbooks[0]
		e.SourceLang = v
		return e
	}
	withTargetLang := func(v string) seed.CSVWorkbookEntry {
		e := wordFillManifest().Workbooks[0]
		e.TargetLang = v
		return e
	}

	tests := []struct {
		name  string
		entry seed.CSVWorkbookEntry
	}{
		{name: "language empty", entry: withLanguage("")},
		{name: "language uppercase", entry: withLanguage("JA")},
		{name: "language three letters", entry: withLanguage("jpn")},
		{name: "sourceLang empty", entry: withSourceLang("")},
		{name: "sourceLang whitespace", entry: withSourceLang("  ")},
		{name: "sourceLang uppercase", entry: withSourceLang("JA")},
		{name: "sourceLang region subtag", entry: withSourceLang("ja-JP")},
		{name: "targetLang empty", entry: withTargetLang("")},
		{name: "targetLang uppercase", entry: withTargetLang("EN")},
		{name: "targetLang three letters", entry: withTargetLang("eng")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// given
			manifest := seed.CSVWorkbookManifest{Workbooks: []seed.CSVWorkbookEntry{tt.entry}}

			// when
			err := seed.ValidateCSVManifest(manifest)

			// then
			require.ErrorIs(t, err, seed.ErrInvalidManifest)
		})
	}
}

func Test_DefaultCSVManifest_shouldNotShareSeedKeyWithDefaultSeeds_whenBothEmbedded(t *testing.T) {
	t.Parallel()

	// given: both kinds of workbooks live in the same public space, keyed by seedKey
	seeds, err := seed.DefaultSeeds()
	require.NoError(t, err)
	yamlKeys := make(map[string]bool, len(seeds))
	for _, s := range seeds {
		yamlKeys[s.SeedKey] = true
	}

	// when
	manifest, err := seed.DefaultCSVManifest()

	// then
	require.NoError(t, err)
	var shared []string
	for _, e := range manifest.Workbooks {
		if yamlKeys[e.SeedKey] {
			shared = append(shared, e.SeedKey)
		}
	}
	assert.Empty(t, shared)
}

func Test_DefaultCSVManifest_shouldMapCEFRB1WorkbookToGeneratedCSV(t *testing.T) {
	t.Parallel()

	// when
	manifest, err := seed.DefaultCSVManifest()

	// then
	require.NoError(t, err)
	gcsObjects := make(map[string]string, len(manifest.Workbooks))
	for _, e := range manifest.Workbooks {
		gcsObjects[e.SeedKey] = e.GCSObject
	}
	assert.Equal(t, "cefr_b1_questions.csv", gcsObjects["cefr-b1-wordfill-v1"])
}
