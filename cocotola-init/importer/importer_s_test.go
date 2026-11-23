//go:build small

package importer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/importer"
	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

const (
	testOrganizationID = "00000000-0000-7000-8000-000000000100"
	testSpaceID        = "00000000-0000-7000-8000-000000000200"
	testSeedKey        = "cefr-b1-wordfill-v1"
)

func loadedWorkbook(seedKey string, invalidRowIndexes ...int32) seed.CSVWorkbookSeed {
	return seed.CSVWorkbookSeed{
		Seed:              seed.PublicWorkbookSeed{SeedKey: seedKey, Title: seedKey},
		InvalidRowIndexes: invalidRowIndexes,
	}
}

func entryFor(seedKey string) seed.CSVWorkbookEntry {
	return seed.CSVWorkbookEntry{SeedKey: seedKey}
}

// indexRange returns the positions first..last.
func indexRange(first, last int32) []int32 {
	indexes := make([]int32, 0, last-first+1)
	for i := first; i <= last; i++ {
		indexes = append(indexes, i)
	}
	return indexes
}

func newFinder(t *testing.T) *MockPublicSpaceFinder {
	t.Helper()
	finder := NewMockPublicSpaceFinder(t)
	finder.EXPECT().FindPublicSpaceID(mock.Anything, testOrganizationID).Return(testSpaceID, nil)
	return finder
}

// newImporter wires mocks so that the space exists, the loader returns each of
// loaded for its entry and the seeder returns the matching element of outputs
// (nothing when outputs is shorter).
func newImporter(t *testing.T, loaded []seed.CSVWorkbookSeed, outputs []seed.WorkbookOutput) *importer.Importer {
	t.Helper()
	entries := make([]seed.CSVWorkbookEntry, 0, len(loaded))
	for _, l := range loaded {
		entries = append(entries, entryFor(l.Seed.SeedKey))
	}
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return(entries)
	seeder := NewMockWorkbookSeeder(t)
	for i, l := range loaded {
		loader.EXPECT().Load(mock.Anything, entries[i]).Return(l, nil)
		var out []seed.WorkbookOutput
		if i < len(outputs) {
			out = []seed.WorkbookOutput{outputs[i]}
		}
		seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, []seed.PublicWorkbookSeed{l.Seed}).Return(out, nil)
	}
	return importer.NewImporter(testOrganizationID, newFinder(t), loader, seeder)
}

func Test_Importer_Run_shouldSeedLoadedWorkbooksIntoPublicSpace_whenSpaceExists(t *testing.T) {
	t.Parallel()

	// given
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, AddedOrderIndexes: indexRange(1, 3)}})

	// when
	err := im.Run(context.Background())

	// then
	require.NoError(t, err)
}

func Test_Importer_Run_shouldReturnError_whenPublicSpaceLookupFails(t *testing.T) {
	t.Parallel()

	// given: loader and seeder must not be called
	findErr := errors.New("public space not found")
	finder := NewMockPublicSpaceFinder(t)
	finder.EXPECT().FindPublicSpaceID(mock.Anything, mock.Anything).Return("", findErr)
	im := importer.NewImporter(testOrganizationID, finder, NewMockCSVWorkbookLoader(t), NewMockWorkbookSeeder(t))

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, findErr)
}

func Test_Importer_Run_shouldReturnError_whenLoadFails(t *testing.T) {
	t.Parallel()

	// given: the seeder must not be called
	loadErr := errors.New("object not found")
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return([]seed.CSVWorkbookEntry{entryFor(testSeedKey)})
	loader.EXPECT().Load(mock.Anything, entryFor(testSeedKey)).Return(seed.CSVWorkbookSeed{}, loadErr)
	im := importer.NewImporter(testOrganizationID, newFinder(t), loader, NewMockWorkbookSeeder(t))

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, loadErr)
}

func Test_Importer_Run_shouldReturnError_whenSeedFails(t *testing.T) {
	t.Parallel()

	// given
	seedErr := errors.New("status 500")
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return([]seed.CSVWorkbookEntry{entryFor(testSeedKey)})
	loader.EXPECT().Load(mock.Anything, entryFor(testSeedKey)).Return(loadedWorkbook(testSeedKey), nil)
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, seedErr)
	im := importer.NewImporter(testOrganizationID, newFinder(t), loader, seeder)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, seedErr)
}

func Test_Importer_Run_shouldReturnErrWorkbookOutputMismatch_whenSeederReturnsNoOutput(t *testing.T) {
	t.Parallel()

	// given
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey)}, nil)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrWorkbookOutputMismatch)
}

func Test_Importer_Run_shouldSeedSecondWorkbook_whenFirstWorkbookLoadFails(t *testing.T) {
	t.Parallel()

	// given
	loadErr := errors.New("object not found")
	second := loadedWorkbook("second")
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return([]seed.CSVWorkbookEntry{entryFor("first"), entryFor("second")})
	loader.EXPECT().Load(mock.Anything, entryFor("first")).Return(seed.CSVWorkbookSeed{}, loadErr)
	loader.EXPECT().Load(mock.Anything, entryFor("second")).Return(second, nil)
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, []seed.PublicWorkbookSeed{second.Seed}).
		Return([]seed.WorkbookOutput{{SeedKey: "second", AddedOrderIndexes: indexRange(1, 3)}}, nil)
	im := importer.NewImporter(testOrganizationID, newFinder(t), loader, seeder)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, loadErr)
}

func Test_Importer_Run_shouldSeedSecondWorkbook_whenFirstWorkbookSeedFails(t *testing.T) {
	t.Parallel()

	// given
	seedErr := errors.New("status 500")
	first := loadedWorkbook("first")
	second := loadedWorkbook("second")
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return([]seed.CSVWorkbookEntry{entryFor("first"), entryFor("second")})
	loader.EXPECT().Load(mock.Anything, entryFor("first")).Return(first, nil)
	loader.EXPECT().Load(mock.Anything, entryFor("second")).Return(second, nil)
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, []seed.PublicWorkbookSeed{first.Seed}).Return(nil, seedErr)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, []seed.PublicWorkbookSeed{second.Seed}).
		Return([]seed.WorkbookOutput{{SeedKey: "second", AddedOrderIndexes: indexRange(1, 3)}}, nil)
	im := importer.NewImporter(testOrganizationID, newFinder(t), loader, seeder)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, seedErr)
}

func Test_Importer_Run_shouldSeedSecondWorkbook_whenFirstWorkbookExceedsLimit(t *testing.T) {
	t.Parallel()

	// given: the first workbook has 2 rejected of 20 new rows
	im := newImporter(t,
		[]seed.CSVWorkbookSeed{loadedWorkbook("first"), loadedWorkbook("second")},
		[]seed.WorkbookOutput{
			{SeedKey: "first", AddedOrderIndexes: indexRange(1, 18), RejectedOrderIndexes: []int32{19, 20}},
			{SeedKey: "second", AddedOrderIndexes: indexRange(1, 3)},
		})

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrTooManyInvalidNewRows)
}

func Test_Importer_Run_shouldCheckSecondWorkbookNewRows_whenFirstWorkbookLoadFails(t *testing.T) {
	t.Parallel()

	// given: the second workbook has 2 rejected of 20 new rows
	loadErr := errors.New("object not found")
	second := loadedWorkbook("second")
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return([]seed.CSVWorkbookEntry{entryFor("first"), entryFor("second")})
	loader.EXPECT().Load(mock.Anything, entryFor("first")).Return(seed.CSVWorkbookSeed{}, loadErr)
	loader.EXPECT().Load(mock.Anything, entryFor("second")).Return(second, nil)
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, []seed.PublicWorkbookSeed{second.Seed}).
		Return([]seed.WorkbookOutput{{SeedKey: "second", AddedOrderIndexes: indexRange(1, 18), RejectedOrderIndexes: []int32{19, 20}}}, nil)
	im := importer.NewImporter(testOrganizationID, newFinder(t), loader, seeder)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrTooManyInvalidNewRows)
}

func Test_Importer_Run_shouldReturnAllErrors_whenTwoWorkbooksFail(t *testing.T) {
	t.Parallel()

	// given
	firstErr := errors.New("object not found")
	secondErr := errors.New("status 500")
	second := loadedWorkbook("second")
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Entries().Return([]seed.CSVWorkbookEntry{entryFor("first"), entryFor("second")})
	loader.EXPECT().Load(mock.Anything, entryFor("first")).Return(seed.CSVWorkbookSeed{}, firstErr)
	loader.EXPECT().Load(mock.Anything, entryFor("second")).Return(second, nil)
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, []seed.PublicWorkbookSeed{second.Seed}).Return(nil, secondErr)
	im := importer.NewImporter(testOrganizationID, newFinder(t), loader, seeder)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, firstErr)
	require.ErrorIs(t, err, secondErr)
}

func Test_Importer_Run_shouldReturnErrTooManyInvalidNewRows_whenNewRowsExceedLimit(t *testing.T) {
	t.Parallel()

	// given: rows 11 and 12 are invalid after the last import at row 10 (2 of 20)
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey, 11, 12)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, AddedOrderIndexes: indexRange(13, 30), SkippedExisting: 10, MaxExistingOrderIndex: 10}})

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrTooManyInvalidNewRows)
}

func Test_Importer_Run_shouldReturnErrTooManyInvalidNewRows_whenRejectedNewRowsExceedLimit(t *testing.T) {
	t.Parallel()

	// given: 2 of 20 new rows were rejected by the server
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, AddedOrderIndexes: indexRange(1, 18), RejectedOrderIndexes: []int32{19, 20}}})

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrTooManyInvalidNewRows)
}

func Test_Importer_Run_shouldReturnErrTooManyInvalidNewRows_whenOnlySecondWorkbookExceedsLimit(t *testing.T) {
	t.Parallel()

	// given: the first workbook is clean; the second has 2 invalid of 20 new rows
	im := newImporter(t,
		[]seed.CSVWorkbookSeed{loadedWorkbook("first"), loadedWorkbook("second", 11, 12)},
		[]seed.WorkbookOutput{
			{SeedKey: "first", AddedOrderIndexes: indexRange(1, 20)},
			{SeedKey: "second", AddedOrderIndexes: indexRange(13, 30), MaxExistingOrderIndex: 10},
		})

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrTooManyInvalidNewRows)
}

func Test_Importer_Run_shouldSucceed_whenInvalidNewRowsAtLimit(t *testing.T) {
	t.Parallel()

	// given: 1 of 20 new rows is invalid (5%)
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey, 11)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, AddedOrderIndexes: indexRange(12, 30), MaxExistingOrderIndex: 10}})

	// when
	err := im.Run(context.Background())

	// then
	require.NoError(t, err)
}

func Test_Importer_Run_shouldIgnoreInvalidRows_whenTheyPrecedeLastImport(t *testing.T) {
	t.Parallel()

	// given: rows 3-5 were invalid before the last import at row 10; one new row
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey, 3, 4, 5)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, AddedOrderIndexes: []int32{11}, SkippedExisting: 7, MaxExistingOrderIndex: 10}})

	// when
	err := im.Run(context.Background())

	// then
	require.NoError(t, err)
}

func Test_Importer_Run_shouldIgnoreRejectedRows_whenTheyPrecedeLastImport(t *testing.T) {
	t.Parallel()

	// given: row 3 is rejected again on every run; one new row after row 10
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, AddedOrderIndexes: []int32{11}, RejectedOrderIndexes: []int32{3}, SkippedExisting: 9, MaxExistingOrderIndex: 10}})

	// when
	err := im.Run(context.Background())

	// then
	require.NoError(t, err)
}

func Test_Importer_Run_shouldSucceed_whenNoNewRows(t *testing.T) {
	t.Parallel()

	// given: nothing appended since the last import at row 20
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey, 3)},
		[]seed.WorkbookOutput{{SeedKey: testSeedKey, SkippedExisting: 19, MaxExistingOrderIndex: 20}})

	// when
	err := im.Run(context.Background())

	// then
	require.NoError(t, err)
}
