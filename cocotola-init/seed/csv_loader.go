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

// Entries returns the manifest entries in manifest order.
func (l *CSVWorkbookLoader) Entries() []CSVWorkbookEntry {
	return l.manifest.Workbooks
}

// Load calls LoadCSVWorkbookSeed with the loader's reader.
func (l *CSVWorkbookLoader) Load(ctx context.Context, entry CSVWorkbookEntry) (CSVWorkbookSeed, error) {
	return LoadCSVWorkbookSeed(ctx, l.reader, entry)
}

// LoadCSVWorkbookSeed downloads the CSV of entry from GCS and converts the rows
// into question seeds. The workbook metadata comes from the entry; only the
// questions come from the CSV. The manifest is validated where it is built.
func LoadCSVWorkbookSeed(ctx context.Context, reader GCSObjectReader, entry CSVWorkbookEntry) (CSVWorkbookSeed, error) {
	data, err := reader.ReadObject(ctx, entry.GCSObject)
	if err != nil {
		return CSVWorkbookSeed{}, fmt.Errorf("read csv object %q for workbook %q: %w", entry.GCSObject, entry.SeedKey, err)
	}

	conv, err := convertCSV(entry.Format, entry.SourceLang, entry.TargetLang, data)
	if err != nil {
		return CSVWorkbookSeed{}, fmt.Errorf("convert csv for workbook %q: %w", entry.SeedKey, err)
	}
	logSkippedRows(ctx, entry.SeedKey, conv.skipped)

	return CSVWorkbookSeed{
		Seed: PublicWorkbookSeed{
			SeedKey:     entry.SeedKey,
			Title:       entry.Title,
			Description: entry.Description,
			Language:    entry.Language,
			Questions:   conv.questions,
		},
		InvalidRowIndexes: invalidRowIndexes(conv.skipped),
	}, nil
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
