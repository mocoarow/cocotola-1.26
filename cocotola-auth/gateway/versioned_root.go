package gateway

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/gormsave"
)

// saveVersionedRoot saves the version row of a multi-row aggregate; a concurrent first insert is reported as a concurrent modification.
func saveVersionedRoot[R libversioned.Record](ctx context.Context, tx *gorm.DB, rootTable string, entity libversioned.Entity, record R, pk map[string]any) error {
	err := gormsave.SaveVersioned(ctx, gormsave.SaveArgs[R]{
		DB:           tx.Table(rootTable),
		Entity:       entity,
		Record:       record,
		PK:           pk,
		Updates:      map[string]any{"updated_at": gorm.Expr("CURRENT_TIMESTAMP")},
		EntityName:   rootTable,
		OmitOnInsert: nil,
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("insert %s: %w", rootTable, libversioned.ErrConcurrentModification)
	}
	if err != nil {
		return fmt.Errorf("save %s: %w", rootTable, err)
	}
	return nil
}
