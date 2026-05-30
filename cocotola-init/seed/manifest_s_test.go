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
