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

// VersionColumn is the version column name every table saved by SaveVersioned must use.
const VersionColumn = "version"

// columnNamePattern guards against SQL injection: PK/Updates keys are
// concatenated into raw SQL.
var columnNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func validateColumnNames(entityName string, m map[string]any) error {
	for col := range m {
		if !columnNamePattern.MatchString(col) {
			return fmt.Errorf("save %s: invalid column name %q", entityName, col)
		}
	}
	return nil
}

func hasNonVersionColumn(m map[string]any) bool {
	for col := range m {
		if col != VersionColumn {
			return true
		}
	}
	return false
}

// SaveArgs bundles the arguments to SaveVersioned.
type SaveArgs[R versioned.Record] struct {
	DB     *gorm.DB
	Entity versioned.Entity
	// Record must already carry version Entity.Version()+1.
	Record R
	PK     map[string]any
	// Updates is used only on update; the version column is always overridden.
	Updates map[string]any
	// EntityName is used only in error messages.
	EntityName string
	// OmitOnInsert is forwarded to GORM Omit on insert only.
	OmitOnInsert []string
}

// SaveVersioned inserts when Entity.Version() is 0, otherwise updates the row
// matching PK and the current version (optimistic lock). Updates must contain
// a non-version column so a forgotten column set cannot bump only the version.
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
	if !hasNonVersionColumn(args.Updates) {
		return fmt.Errorf("save %s: updates must contain a non-version column on update path", args.EntityName)
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

// classifyCasMiss tells a deleted row from a version mismatch after a
// 0-RowsAffected update.
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
