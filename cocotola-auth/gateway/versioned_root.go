package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
	"github.com/mocoarow/cocotola-1.26/cocotola-lib/gateway/gormsave"
)

const (
	organizationIDColumn = "organization_id"
	groupIDColumn        = "group_id"
	userIDColumn         = "user_id"
)

type organizationRootRecord struct {
	OrganizationID string    `gorm:"column:organization_id;primaryKey"`
	Version        int       `gorm:"column:version"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	UpdatedAt      time.Time `gorm:"column:updated_at;->"`
}

func (r *organizationRootRecord) GetVersion() int { return r.Version }

func newOrganizationRootRecord(organizationID string, entity libversioned.Entity) *organizationRootRecord {
	return &organizationRootRecord{OrganizationID: organizationID, Version: entity.Version() + 1, CreatedAt: time.Time{}, UpdatedAt: time.Time{}}
}

type groupRootRecord struct {
	GroupID   string    `gorm:"column:group_id;primaryKey"`
	Version   int       `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at;->"`
	UpdatedAt time.Time `gorm:"column:updated_at;->"`
}

func (r *groupRootRecord) GetVersion() int { return r.Version }

func newGroupRootRecord(groupID string, entity libversioned.Entity) *groupRootRecord {
	return &groupRootRecord{GroupID: groupID, Version: entity.Version() + 1, CreatedAt: time.Time{}, UpdatedAt: time.Time{}}
}

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

// findRootVersion returns the version stored in the root row keyed by keyColumn, or 0 when the row does not exist.
func findRootVersion(ctx context.Context, db *gorm.DB, rootTable string, keyColumn string, key string) (int, error) {
	var versions []int
	if err := db.WithContext(ctx).Table(rootTable).Where(keyColumn+" = ?", key).Pluck(gormsave.VersionColumn, &versions).Error; err != nil {
		return 0, fmt.Errorf("find %s by %s %s: %w", rootTable, keyColumn, key, err)
	}
	if len(versions) == 0 {
		return 0, nil
	}
	return versions[0], nil
}

// saveVersionedMembers saves the root row and replaces the member rows keyed by keyColumn in one transaction.
func saveVersionedMembers[R libversioned.Record, M schema.Tabler](ctx context.Context, db *gorm.DB, rootTable string, entity libversioned.Entity, root R, keyColumn string, key string, members []M) error {
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := saveVersionedRoot(ctx, tx, rootTable, entity, root, map[string]any{keyColumn: key}); err != nil {
			return err
		}
		if err := tx.Where(keyColumn+" = ?", key).Delete(new(M)).Error; err != nil {
			return fmt.Errorf("delete members: %w", err)
		}
		if len(members) == 0 {
			return nil
		}
		if err := tx.Create(&members).Error; err != nil {
			return fmt.Errorf("insert members: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("save %s with members: %w", rootTable, err)
	}
	return nil
}
