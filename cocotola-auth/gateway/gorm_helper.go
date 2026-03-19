package gateway

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// replaceRecords deletes all records matching the where clause, then inserts new records,
// all within a single transaction.
// It first queries existing records via snapshot read (no locks), then deletes only if
// records exist, using primary-key-based deletion to avoid InnoDB gap locks.
func replaceRecords[R any](ctx context.Context, db *gorm.DB, whereClause string, whereArg any, newRecords []R, label string) error {
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []R
		if err := tx.Where(whereClause, whereArg).Find(&existing).Error; err != nil {
			return fmt.Errorf("find existing %s: %w", label, err)
		}
		if len(existing) > 0 {
			if err := tx.Delete(&existing).Error; err != nil {
				return fmt.Errorf("delete %s: %w", label, err)
			}
		}
		if len(newRecords) == 0 {
			return nil
		}
		if err := tx.Create(&newRecords).Error; err != nil {
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
func findAndConvertWhitelist[R any](ctx context.Context, db *gorm.DB, userID int, toEntry func(R) domain.WhitelistEntry, label string) ([]domain.WhitelistEntry, error) {
	var records []R
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s: %w", label, err)
	}

	entries := make([]domain.WhitelistEntry, len(records))
	for i := range records {
		entries[i] = toEntry(records[i])
	}
	return entries, nil
}

// saveWhitelist converts domain whitelist entries to records and persists them.
func saveWhitelist[R any](ctx context.Context, db *gorm.DB, whitelist *domain.TokenWhitelist, toRecord func(int, domain.WhitelistEntry) R, label string) error {
	entries := whitelist.Entries()

	records := make([]R, len(entries))
	for i, e := range entries {
		records[i] = toRecord(whitelist.UserID(), e)
	}
	return replaceRecords(ctx, db, "user_id = ?", whitelist.UserID(), records, label)
}

// findMemberIDs queries records by organization_id and extracts member IDs.
func findMemberIDs[R any](ctx context.Context, db *gorm.DB, organizationID int, extractID func(R) int, label string) ([]int, error) {
	var records []R
	if err := db.WithContext(ctx).Where("organization_id = ?", organizationID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s: %w", label, err)
	}

	ids := make([]int, len(records))
	for i := range records {
		ids[i] = extractID(records[i])
	}
	return ids, nil
}
