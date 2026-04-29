package gateway

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaintoken "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/token"
)

// replaceRecords deletes existing records matching the where clause and inserts
// new records, all within a single transaction.
//
// Concurrency:
//   - The SELECT acquires `FOR UPDATE` locks so concurrent transactions modifying
//     the same key set are serialized; the second transaction blocks until the
//     first commits, then sees the committed state.
//   - The INSERT uses `ON CONFLICT DO NOTHING` so that any leftover duplicate
//     primary keys (e.g. created by a concurrent re-insert that committed
//     between this transaction's DELETE and INSERT) do not cause a unique-key
//     violation; the existing row is kept and the next save will reconcile.
//
// Without these, the delete-then-insert pattern races under concurrent same-key
// writes: two transactions can both read the same snapshot, both delete the
// same rows, and then collide on INSERT once one of them commits before the
// other reaches its INSERT statement.
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

// findRecordByHash looks up a record by token_hash column and returns ErrTokenNotFound
// when no matching record exists.
func findRecordByHash[R any](ctx context.Context, db *gorm.DB, hash string, label string) (*R, error) {
	var record R
	if err := db.WithContext(ctx).Where("token_hash = ?", hash).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTokenNotFound
		}
		return nil, fmt.Errorf("find %s by hash: %w", label, err)
	}
	return &record, nil
}

// findAndConvertWhitelist queries whitelist records and converts them to domain entries.
func findAndConvertWhitelist[R any](ctx context.Context, db *gorm.DB, userID domain.AppUserID, toEntry func(R) domaintoken.WhitelistEntry, label string) ([]domaintoken.WhitelistEntry, error) {
	var records []R
	if err := db.WithContext(ctx).Where("user_id = ?", userID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s: %w", label, err)
	}

	entries := make([]domaintoken.WhitelistEntry, len(records))
	for i := range records {
		entries[i] = toEntry(records[i])
	}
	return entries, nil
}

// saveWhitelist converts domain whitelist entries to records and persists them.
func saveWhitelist[R any](ctx context.Context, db *gorm.DB, whitelist *domaintoken.Whitelist, toRecord func(string, domaintoken.WhitelistEntry) R, label string) error {
	entries := whitelist.Entries()
	userIDStr := whitelist.UserID().String()

	records := make([]R, len(entries))
	for i, e := range entries {
		records[i] = toRecord(userIDStr, e)
	}
	return replaceRecords(ctx, db, "user_id = ?", userIDStr, records, label)
}

// findMemberIDs queries records by organization_id and extracts member IDs as strings.
func findMemberIDs[R any, ID any](ctx context.Context, db *gorm.DB, organizationID domain.OrganizationID, extractID func(R) ID, label string) ([]ID, error) {
	var records []R
	if err := db.WithContext(ctx).Where("organization_id = ?", organizationID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s: %w", label, err)
	}

	ids := make([]ID, len(records))
	for i := range records {
		ids[i] = extractID(records[i])
	}
	return ids, nil
}
