package gormsave

import (
	"context"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

// VersionColumn is the database column name used by SaveVersioned to perform
// the optimistic-lock compare-and-swap. All tables managed by this helper must
// use the same column name for the version field.
const VersionColumn = "version"

// columnNamePattern validates SQL identifier names that callers pass through
// pk and updates. Keys are concatenated into raw SQL ("col = ?") and used as
// GORM Updates map keys; allowing arbitrary strings would expose SQL injection
// when callers forward untrusted input. We accept the canonical unquoted
// identifier form (letter or underscore, then letters/digits/underscores).
var columnNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateColumnNames(entityName string, m map[string]any) error {
	for col := range m {
		if !columnNamePattern.MatchString(col) {
			return fmt.Errorf("save %s: invalid column name %q", entityName, col)
		}
	}
	return nil
}

// SaveArgs bundles the arguments to SaveVersioned. R is the persistence record
// type (typically a pointer to a GORM-mapped struct) constrained by
// versioned.Record so the helper can verify the record's version without
// reflection.
type SaveArgs[R versioned.Record] struct {
	// DB is the GORM session used for the insert/update.
	DB *gorm.DB
	// Entity is the aggregate whose version drives the CAS.
	Entity versioned.Entity
	// Record is the row to insert (entity.Version()==0) or whose values to set
	// on update (entity.Version()>0). The caller must set Record.GetVersion()
	// to Entity.Version()+1 before calling SaveVersioned.
	Record R
	// PK is the primary-key columns used in the WHERE clause on update. Must
	// be non-empty; keys must match [A-Za-z_][A-Za-z0-9_]*.
	PK map[string]any
	// Updates is the column set to write on update. Keys must match
	// [A-Za-z_][A-Za-z0-9_]*. The helper appends VersionColumn=next-version,
	// overriding any version key the caller may have included.
	Updates map[string]any
	// EntityName is interpolated into error messages (e.g. "app user").
	EntityName string
	// OmitOnInsert lists columns to exclude on the insert path (forwarded to
	// GORM .Omit). Ignored on the update path.
	OmitOnInsert []string
}

// SaveVersioned persists a versioned aggregate using GORM with optimistic
// concurrency control.
//
// When Entity.Version() == 0 the record is inserted. When Entity.Version() > 0
// the row identified by PK is updated, with WHERE clauses on each PK column
// AND the current version. Updates contains the columns to set; the helper
// adds VersionColumn = Entity.Version()+1 to the update set, overriding any
// version key the caller may have included.
//
// On a successful update with zero affected rows the helper distinguishes
// between two cases by re-querying with the PK only: if no row exists it
// returns versioned.ErrNotFound; otherwise the row's version no longer
// matches and it returns versioned.ErrConcurrentModification. Callers should
// translate ErrNotFound into their domain-specific not-found error and treat
// ErrConcurrentModification as a CAS miss to reload and retry.
//
// The caller MUST set the record's version to Entity.Version()+1 before
// calling this helper. The helper verifies this via the Record interface and
// returns an error if it does not match; this guards against a forgotten
// version assignment silently writing a stale version on the insert path. The
// helper does not mutate fields on the supplied record. On any successful
// save the helper calls Entity.SetVersion with the new version.
//
// PK and Updates keys must be SQL identifiers matching [A-Za-z_][A-Za-z0-9_]*.
// They are concatenated into raw SQL fragments and used as GORM Updates map
// keys; the helper rejects keys that do not match this pattern to prevent
// injection. Values are always passed as bound parameters. WHERE clauses are
// emitted in lexicographic order of PK keys so generated SQL is stable across
// runs.
func SaveVersioned[R versioned.Record](ctx context.Context, args SaveArgs[R]) error {
	if len(args.PK) == 0 {
		return fmt.Errorf("save %s: pk must not be empty", args.EntityName)
	}
	if err := validateColumnNames(args.EntityName, args.PK); err != nil {
		return err
	}
	if err := validateColumnNames(args.EntityName, args.Updates); err != nil {
		return err
	}

	nextVersion := args.Entity.Version() + 1
	if got := args.Record.GetVersion(); got != nextVersion {
		return fmt.Errorf("save %s: record Version=%d does not match expected next version %d", args.EntityName, got, nextVersion)
	}

	if args.Entity.Version() == 0 {
		return insert(ctx, args, nextVersion)
	}
	return update(ctx, args, nextVersion)
}

func insert[R versioned.Record](ctx context.Context, args SaveArgs[R], nextVersion int) error {
	session := args.DB.WithContext(ctx)
	if len(args.OmitOnInsert) > 0 {
		session = session.Omit(args.OmitOnInsert...)
	}
	if err := session.Create(args.Record).Error; err != nil {
		return fmt.Errorf("insert %s: %w", args.EntityName, err)
	}
	args.Entity.SetVersion(nextVersion)
	return nil
}

func update[R versioned.Record](ctx context.Context, args SaveArgs[R], nextVersion int) error {
	pkCols := slices.Sorted(maps.Keys(args.PK))

	updateSet := make(map[string]any, len(args.Updates)+1)
	maps.Copy(updateSet, args.Updates)
	updateSet[VersionColumn] = nextVersion

	session := args.DB.WithContext(ctx).Model(args.Record)
	for _, col := range pkCols {
		session = session.Where(col+" = ?", args.PK[col])
	}
	session = session.Where(VersionColumn+" = ?", args.Entity.Version())

	result := session.Updates(updateSet)
	if result.Error != nil {
		return fmt.Errorf("update %s: %w", args.EntityName, result.Error)
	}
	if result.RowsAffected == 0 {
		return classifyCasMiss(ctx, args, pkCols)
	}
	args.Entity.SetVersion(nextVersion)
	return nil
}

// classifyCasMiss disambiguates a 0-RowsAffected update by re-querying with
// the primary key only. If no row exists it returns versioned.ErrNotFound;
// otherwise the version column did not match and it returns
// versioned.ErrConcurrentModification.
func classifyCasMiss[R versioned.Record](ctx context.Context, args SaveArgs[R], pkCols []string) error {
	var count int64
	session := args.DB.WithContext(ctx).Model(args.Record)
	for _, col := range pkCols {
		session = session.Where(col+" = ?", args.PK[col])
	}
	if err := session.Count(&count).Error; err != nil {
		return fmt.Errorf("classify cas miss for %s: %w", args.EntityName, err)
	}
	if count == 0 {
		return versioned.ErrNotFound
	}
	return versioned.ErrConcurrentModification
}
