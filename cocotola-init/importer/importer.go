package importer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"

	"github.com/mocoarow/cocotola-1.26/cocotola-init/seed"
)

// maxInvalidNewRowPercent is the share of newly appended rows that may be
// invalid or rejected before an import fails.
const maxInvalidNewRowPercent = 5

var (
	// ErrTooManyInvalidNewRows is returned when more than
	// maxInvalidNewRowPercent of the rows appended since the last import are
	// invalid or rejected.
	ErrTooManyInvalidNewRows = errors.New("too many invalid new rows")
	// ErrWorkbookOutputMismatch is returned when the seeder does not return one
	// output per loaded workbook.
	ErrWorkbookOutputMismatch = errors.New("workbook output mismatch")
)

// PublicSpaceFinder finds an organization's public space.
type PublicSpaceFinder interface {
	FindPublicSpaceID(ctx context.Context, organizationID domain.OrganizationID) (domain.SpaceID, error)
}

// CSVWorkbookLoader loads the workbooks to import.
type CSVWorkbookLoader interface {
	Load(ctx context.Context) ([]seed.CSVWorkbookSeed, error)
}

// WorkbookSeeder adds missing workbooks and questions.
type WorkbookSeeder interface {
	SeedWorkbooks(ctx context.Context, organizationID, publicSpaceID string, seeds []seed.PublicWorkbookSeed) ([]seed.WorkbookOutput, error)
}

// Importer imports CSV workbooks into an organization's public space.
type Importer struct {
	organizationID domain.OrganizationID
	finder         PublicSpaceFinder
	loader         CSVWorkbookLoader
	seeder         WorkbookSeeder
	tracer         trace.Tracer
	logger         *slog.Logger
}

// NewImporter returns an Importer for organizationID.
func NewImporter(organizationID domain.OrganizationID, finder PublicSpaceFinder, loader CSVWorkbookLoader, seeder WorkbookSeeder) *Importer {
	return &Importer{
		organizationID: organizationID,
		finder:         finder,
		loader:         loader,
		seeder:         seeder,
		tracer:         otel.Tracer("github.com/mocoarow/cocotola-1.26/cocotola-init/importer"),
		logger:         slog.Default().With(slog.String("component", "csv-importer")),
	}
}

// Run imports the CSV workbooks and fails when too many of the newly appended
// rows are invalid or rejected.
func (im *Importer) Run(ctx context.Context) error {
	ctx, span := im.tracer.Start(ctx, "importer.Run")
	defer span.End()

	if err := im.run(ctx, span); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "import failed")
		return err
	}
	return nil
}

func (im *Importer) run(ctx context.Context, span trace.Span) error {
	spaceID, err := im.finder.FindPublicSpaceID(ctx, im.organizationID)
	if err != nil {
		return fmt.Errorf("find public space: %w", err)
	}

	loaded, err := im.loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("load csv workbooks: %w", err)
	}

	seeds := make([]seed.PublicWorkbookSeed, 0, len(loaded))
	for _, l := range loaded {
		seeds = append(seeds, l.Seed)
	}
	outputs, err := im.seeder.SeedWorkbooks(ctx, im.organizationID.String(), spaceID.String(), seeds)
	if err != nil {
		return fmt.Errorf("seed workbooks: %w", err)
	}
	if len(outputs) != len(loaded) {
		return fmt.Errorf("%d outputs for %d workbooks: %w", len(outputs), len(loaded), ErrWorkbookOutputMismatch)
	}

	for i, output := range outputs {
		if err := im.checkNewRows(ctx, span, loaded[i].InvalidRowIndexes, output); err != nil {
			return err
		}
	}
	return nil
}

// checkNewRows judges only the rows after the last imported one, so rows that
// were invalid in earlier imports do not hide a broken generator today.
func (im *Importer) checkNewRows(ctx context.Context, span trace.Span, invalidRowIndexes []int32, output seed.WorkbookOutput) error {
	after := output.MaxExistingOrderIndex
	invalid := countAfter(invalidRowIndexes, after)
	rejected := countAfter(output.RejectedOrderIndexes, after)
	added := countAfter(output.AddedOrderIndexes, after)
	bad := invalid + rejected
	newRows := bad + added

	span.AddEvent("workbook imported", trace.WithAttributes(
		attribute.String("seed_key", output.SeedKey),
		attribute.Int("new_rows", newRows),
		attribute.Int("added", added),
		attribute.Int("invalid", invalid),
		attribute.Int("rejected", rejected),
	))
	im.logger.InfoContext(ctx, "workbook imported",
		slog.String("seed_key", output.SeedKey),
		slog.Int("new_rows", newRows),
		slog.Int("added", added),
		slog.Int("invalid", invalid),
		slog.Int("rejected", rejected),
	)

	if bad*100 > newRows*maxInvalidNewRowPercent {
		return fmt.Errorf("workbook %q: %d of %d new rows are invalid or rejected: %w", output.SeedKey, bad, newRows, ErrTooManyInvalidNewRows)
	}
	return nil
}

func countAfter(indexes []int32, after int32) int {
	count := 0
	for _, index := range indexes {
		if index > after {
			count++
		}
	}
	return count
}
