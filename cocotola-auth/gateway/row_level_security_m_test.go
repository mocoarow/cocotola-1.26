//go:build medium

package gateway_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_PublicTables_shouldEnableRowLevelSecurity_whenMigrated(t *testing.T) {
	t.Parallel()

	// given
	var tablesWithoutRLS []string

	// when
	err := testDB.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND NOT rowsecurity ORDER BY tablename").Scan(&tablesWithoutRLS).Error

	// then
	require.NoError(t, err)
	assert.Empty(t, tablesWithoutRLS)
}
