package gateway

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// replaceRecords deletes matching records and inserts newRecords in one transaction.
// FOR UPDATE serializes concurrent same-key writers; ON CONFLICT DO NOTHING absorbs
// a row re-inserted concurrently between DELETE and INSERT (the next save reconciles).
func replaceRecords[R any](ctx context.Context, db *gorm.DB, whereClause string, whereArg any, newRecords []R, label string) error {
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lockClause := clause.Locking{
			Strength: clause.LockingStrengthUpdate,
			Table:    clause.Table{Name: "", Alias: "", Raw: false},
			Options:  "",
		}
		var existing []R
		if err := tx.Clauses(lockClause).Where(whereClause, whereArg).Find(&existing).Error; err != nil {
			return fmt.Errorf("find existing %s for update: %w", label, err)
		}
		if len(existing) > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return fmt.Errorf("delete %s: %w", label, err)
			}
		}
		if len(newRecords) == 0 {
			return nil
		}
		emptyWhere := clause.Where{Exprs: nil}
		conflictClause := clause.OnConflict{
			Columns:      nil,
			Where:        emptyWhere,
			TargetWhere:  emptyWhere,
			OnConstraint: "",
			DoNothing:    true,
			DoUpdates:    nil,
			UpdateAll:    false,
		}
		if err := tx.Clauses(conflictClause).Create(&newRecords).Error; err != nil {
			return fmt.Errorf("insert %s: %w", label, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("save %s: %w", label, err)
	}
	return nil
}

// findMemberIDs queries records by organization_id and extracts member IDs as strings.
func findMemberIDs[R any, ID any](ctx context.Context, db *gorm.DB, organizationID domain.OrganizationID, extractID func(R) (ID, error), label string) ([]ID, error) {
	var records []R
	if err := db.WithContext(ctx).Where("organization_id = ?", organizationID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s: %w", label, err)
	}

	ids := make([]ID, len(records))
	for i := range records {
		id, err := extractID(records[i])
		if err != nil {
			return nil, fmt.Errorf("extract id from %s record: %w", label, err)
		}
		ids[i] = id
	}
	return ids, nil
}
