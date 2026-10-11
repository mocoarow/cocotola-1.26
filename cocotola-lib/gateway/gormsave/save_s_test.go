package gormsave_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	gormtests "gorm.io/gorm/utils/tests"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/gormsave"
)

type fakeEntity struct {
	version int
}

func (e *fakeEntity) Version() int     { return e.version }
func (e *fakeEntity) SetVersion(v int) { e.version = v }

type fakeRecord struct {
	ID      string
	Version int
}

func (r *fakeRecord) GetVersion() int { return r.Version }

func Test_VersionColumn_shouldBeVersion(t *testing.T) {
	t.Parallel()
	// given
	// when
	got := gormsave.VersionColumn

	// then
	assert.Equal(t, "version", got)
}

func Test_SaveVersioned_shouldReturnError_whenPkIsEmpty(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 0},
		Record:     &fakeRecord{ID: "x", Version: 1},
		PK:         map[string]any{},
		Updates:    map[string]any{"name": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "pk must not be empty")
}

func Test_SaveVersioned_shouldReturnError_whenPkKeyContainsSpace(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 1},
		Record:     &fakeRecord{ID: "x", Version: 2},
		PK:         map[string]any{"id; DROP TABLE users--": "x"},
		Updates:    map[string]any{"name": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "invalid column name")
}

func Test_SaveVersioned_shouldReturnError_whenPkKeyStartsWithDigit(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 1},
		Record:     &fakeRecord{ID: "x", Version: 2},
		PK:         map[string]any{"1id": "x"},
		Updates:    map[string]any{"name": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "invalid column name")
}

func Test_SaveVersioned_shouldReturnError_whenUpdatesKeyContainsQuote(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 1},
		Record:     &fakeRecord{ID: "x", Version: 2},
		PK:         map[string]any{"id": "x"},
		Updates:    map[string]any{"name'": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "invalid column name")
}

func Test_SaveVersioned_shouldReturnError_whenUpdatesKeyContainsHyphen(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 1},
		Record:     &fakeRecord{ID: "x", Version: 2},
		PK:         map[string]any{"id": "x"},
		Updates:    map[string]any{"login-id": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "invalid column name")
}

func Test_SaveVersioned_shouldEmbedEntityNameInError_whenValidationFails(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 1},
		Record:     &fakeRecord{ID: "x", Version: 2},
		PK:         map[string]any{},
		Updates:    map[string]any{"name": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "save app user")
}

func Test_SaveVersioned_shouldReturnError_whenRecordVersionDoesNotMatchNextVersion(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 3},
		Record:     &fakeRecord{ID: "x", Version: 3},
		PK:         map[string]any{"id": "x"},
		Updates:    map[string]any{"name": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "record Version=3 does not match expected next version 4")
}

func Test_SaveVersioned_shouldReturnError_whenRecordVersionIsZeroOnInsert(t *testing.T) {
	t.Parallel()
	// given
	ctx := context.Background()

	// when
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
		DB:         nil,
		Entity:     &fakeEntity{version: 0},
		Record:     &fakeRecord{ID: "x", Version: 0},
		PK:         map[string]any{"id": "x"},
		Updates:    map[string]any{"name": "x"},
		EntityName: "app user",
	})

	// then
	require.ErrorContains(t, err, "record Version=0 does not match expected next version 1")
}

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(gormtests.DummyDialector{}, &gorm.Config{DryRun: true})
	require.NoError(t, err)
	return db
}

func Test_SaveVersioned_shouldReturnError_whenUpdatesHasNoNonVersionColumnOnUpdatePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		updates map[string]any
	}{
		{name: "nil map", updates: nil},
		{name: "empty map", updates: map[string]any{}},
		{name: "version column only", updates: map[string]any{gormsave.VersionColumn: 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// given
			ctx := context.Background()

			// when
			err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
				DB:         nil,
				Entity:     &fakeEntity{version: 3},
				Record:     &fakeRecord{ID: "x", Version: 4},
				PK:         map[string]any{"id": "x"},
				Updates:    tt.updates,
				EntityName: "app user",
			})

			// then
			require.ErrorContains(t, err, "updates must contain a non-version column on update path")
		})
	}
}

func Test_SaveVersioned_shouldReturnNoError_whenUpdatesIsEmptyOnInsertPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		updates map[string]any
	}{
		{name: "nil map", updates: nil},
		{name: "empty map", updates: map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// given
			ctx := context.Background()

			// when
			err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
				DB:         newDryRunDB(t),
				Entity:     &fakeEntity{version: 0},
				Record:     &fakeRecord{ID: "x", Version: 1},
				PK:         map[string]any{"id": "x"},
				Updates:    tt.updates,
				EntityName: "app user",
			})

			// then
			require.NoError(t, err)
		})
	}
}

func Test_SaveVersioned_shouldSetEntityVersion_whenUpdatesIsEmptyOnInsertPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		updates map[string]any
	}{
		{name: "nil map", updates: nil},
		{name: "empty map", updates: map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// given
			ctx := context.Background()
			entity := &fakeEntity{version: 0}

			// when
			err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[*fakeRecord]{
				DB:         newDryRunDB(t),
				Entity:     entity,
				Record:     &fakeRecord{ID: "x", Version: 1},
				PK:         map[string]any{"id": "x"},
				Updates:    tt.updates,
				EntityName: "app user",
			})
			require.NoError(t, err)

			// then
			assert.Equal(t, 1, entity.Version())
		})
	}
}
