//go:build small

package importer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

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

// indexRange returns the positions first..last.
func indexRange(first, last int32) []int32 {
	indexes := make([]int32, 0, last-first+1)
	for i := first; i <= last; i++ {
		indexes = append(indexes, i)
	}
	return indexes
}

// newImporter wires mocks so that the space exists, the loader returns loaded
// and the seeder returns outputs.
func newImporter(t *testing.T, loaded []seed.CSVWorkbookSeed, outputs []seed.WorkbookOutput) *importer.Importer {
	t.Helper()
	finder := NewMockPublicSpaceFinder(t)
	finder.EXPECT().FindPublicSpaceID(mock.Anything, domain.MustParseOrganizationID(testOrganizationID)).
		Return(domain.MustParseSpaceID(testSpaceID), nil)
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Load(mock.Anything).Return(loaded, nil)
	seeds := make([]seed.PublicWorkbookSeed, 0, len(loaded))
	for _, l := range loaded {
		seeds = append(seeds, l.Seed)
	}
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, testOrganizationID, testSpaceID, seeds).Return(outputs, nil)
	return importer.NewImporter(domain.MustParseOrganizationID(testOrganizationID), finder, loader, seeder)
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
	finder.EXPECT().FindPublicSpaceID(mock.Anything, mock.Anything).Return(domain.SpaceID{}, findErr)
	im := importer.NewImporter(domain.MustParseOrganizationID(testOrganizationID), finder, NewMockCSVWorkbookLoader(t), NewMockWorkbookSeeder(t))

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, findErr)
}

func Test_Importer_Run_shouldReturnError_whenLoadFails(t *testing.T) {
	t.Parallel()

	// given: the seeder must not be called
	loadErr := errors.New("object not found")
	finder := NewMockPublicSpaceFinder(t)
	finder.EXPECT().FindPublicSpaceID(mock.Anything, mock.Anything).Return(domain.MustParseSpaceID(testSpaceID), nil)
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Load(mock.Anything).Return(nil, loadErr)
	im := importer.NewImporter(domain.MustParseOrganizationID(testOrganizationID), finder, loader, NewMockWorkbookSeeder(t))

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, loadErr)
}

func Test_Importer_Run_shouldReturnError_whenSeedFails(t *testing.T) {
	t.Parallel()

	// given
	seedErr := errors.New("status 500")
	finder := NewMockPublicSpaceFinder(t)
	finder.EXPECT().FindPublicSpaceID(mock.Anything, mock.Anything).Return(domain.MustParseSpaceID(testSpaceID), nil)
	loader := NewMockCSVWorkbookLoader(t)
	loader.EXPECT().Load(mock.Anything).Return([]seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey)}, nil)
	seeder := NewMockWorkbookSeeder(t)
	seeder.EXPECT().SeedWorkbooks(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, seedErr)
	im := importer.NewImporter(domain.MustParseOrganizationID(testOrganizationID), finder, loader, seeder)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, seedErr)
}

func Test_Importer_Run_shouldReturnErrWorkbookOutputMismatch_whenSeederReturnsFewerOutputs(t *testing.T) {
	t.Parallel()

	// given
	im := newImporter(t, []seed.CSVWorkbookSeed{loadedWorkbook(testSeedKey)}, nil)

	// when
	err := im.Run(context.Background())

	// then
	require.ErrorIs(t, err, importer.ErrWorkbookOutputMismatch)
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
