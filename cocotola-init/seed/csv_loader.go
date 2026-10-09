package seed

import (
	"context"
	"fmt"
	"log/slog"
)

// maxSkippedRowLogs caps the per-row warnings for one CSV; the rest are
// summarized in a single line.
const maxSkippedRowLogs = 10

// GCSObjectReader reads the full bytes of an object identified by its key.
// Declared here (used-side interface) so the CSV loader can be tested without
// a real GCS client.
type GCSObjectReader interface {
	ReadObject(ctx context.Context, objectKey string) ([]byte, error)
}

// CSVWorkbookSeed is a workbook seed loaded from a CSV with the positions of
// the rows skipped as invalid, on the same basis as QuestionSeed.OrderIndex.
type CSVWorkbookSeed struct {
	Seed              PublicWorkbookSeed
	InvalidRowIndexes []int32
}

// CSVWorkbookLoader loads the workbooks declared in a manifest from GCS.
type CSVWorkbookLoader struct {
	reader   GCSObjectReader
	manifest CSVWorkbookManifest
}

// NewCSVWorkbookLoader returns a CSVWorkbookLoader for manifest.
func NewCSVWorkbookLoader(reader GCSObjectReader, manifest CSVWorkbookManifest) *CSVWorkbookLoader {
	return &CSVWorkbookLoader{reader: reader, manifest: manifest}
}

// Load calls LoadCSVWorkbookSeeds with the loader's reader and manifest.
func (l *CSVWorkbookLoader) Load(ctx context.Context) ([]CSVWorkbookSeed, error) {
	return LoadCSVWorkbookSeeds(ctx, l.reader, l.manifest)
}

// LoadCSVWorkbookSeeds turns each manifest entry into a CSVWorkbookSeed by
// downloading its CSV from GCS and converting the rows into question seeds.
// The workbook metadata comes from the manifest; only the questions come from
// the CSV.
func LoadCSVWorkbookSeeds(ctx context.Context, reader GCSObjectReader, manifest CSVWorkbookManifest) ([]CSVWorkbookSeed, error) {
	if err := manifest.validate(); err != nil {
		return nil, fmt.Errorf("validate csv manifest: %w", err)
	}

	seeds := make([]CSVWorkbookSeed, 0, len(manifest.Workbooks))
	for _, entry := range manifest.Workbooks {
		data, err := reader.ReadObject(ctx, entry.GCSObject)
		if err != nil {
			return nil, fmt.Errorf("read csv object %q for workbook %q: %w", entry.GCSObject, entry.SeedKey, err)
		}

		conv, err := convertCSV(entry.Format, entry.SourceLang, entry.TargetLang, data)
		if err != nil {
			return nil, fmt.Errorf("convert csv for workbook %q: %w", entry.SeedKey, err)
		}
		logSkippedRows(ctx, entry.SeedKey, conv.skipped)
		if conv.tooManyInvalidRows() {
			return nil, fmt.Errorf("workbook %q: %d of %d rows are invalid: %w", entry.SeedKey, len(conv.skipped), conv.totalRows, ErrTooManyInvalidCSVRows)
		}

		seeds = append(seeds, CSVWorkbookSeed{
			Seed: PublicWorkbookSeed{
				SeedKey:     entry.SeedKey,
				Title:       entry.Title,
				Description: entry.Description,
				Language:    entry.Language,
				Questions:   conv.questions,
			},
			InvalidRowIndexes: invalidRowIndexes(conv.skipped),
		})
	}

	return seeds, nil
}

func invalidRowIndexes(skipped []skippedCSVRow) []int32 {
	indexes := make([]int32, 0, len(skipped))
	for _, row := range skipped {
		indexes = append(indexes, row.index)
	}
	return indexes
}

func logSkippedRows(ctx context.Context, seedKey string, skipped []skippedCSVRow) {
	for _, row := range skipped[:min(len(skipped), maxSkippedRowLogs)] {
		slog.WarnContext(ctx, "skip invalid csv row",
			slog.String("seed_key", seedKey),
			slog.Int("record", int(row.index)+csvHeaderRows),
			slog.Any("reason", row.reason),
		)
	}
	if omitted := len(skipped) - maxSkippedRowLogs; omitted > 0 {
		slog.WarnContext(ctx, "skip more invalid csv rows",
			slog.String("seed_key", seedKey),
			slog.Int("omitted", omitted),
		)
	}
}
