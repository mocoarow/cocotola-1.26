//go:build small

package seed_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

func Test_MergeSeeds_shouldConcatenateGroups_whenSeedKeysUnique(t *testing.T) {
	t.Parallel()

	// given
	groupA := []seed.PublicWorkbookSeed{{SeedKey: "a", Title: "A", Language: "ja"}}
	groupB := []seed.PublicWorkbookSeed{{SeedKey: "b", Title: "B", Language: "ja"}}

	// when
	merged, err := seed.MergeSeeds(groupA, groupB)

	// then
	require.NoError(t, err)
	require.Len(t, merged, 2)
	assert.Equal(t, "a", merged[0].SeedKey)
	assert.Equal(t, "b", merged[1].SeedKey)
}

func Test_MergeSeeds_shouldReturnError_whenSeedKeyDuplicatedAcrossGroups(t *testing.T) {
	t.Parallel()

	// given: the same seedKey appears in two different groups
	groupA := []seed.PublicWorkbookSeed{{SeedKey: "dup", Title: "A", Language: "ja"}}
	groupB := []seed.PublicWorkbookSeed{{SeedKey: "dup", Title: "B", Language: "ja"}}

	// when
	_, err := seed.MergeSeeds(groupA, groupB)

	// then
	require.ErrorIs(t, err, seed.ErrInvalidSeed)
}
