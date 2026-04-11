package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domainspace "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/space"
)

type spaceRecord struct {
	ID             string    `gorm:"column:id;primaryKey"`
	Version        int       `gorm:"column:version"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	UpdatedAt      time.Time `gorm:"column:updated_at;->"`
	CreatedBy      string    `gorm:"column:created_by;<-:create"`
	UpdatedBy      string    `gorm:"column:updated_by"`
	OrganizationID string    `gorm:"column:organization_id"`
	OwnerID        string    `gorm:"column:owner_id"`
	KeyName        string    `gorm:"column:key_name"`
	Name           string    `gorm:"column:name"`
	SpaceType      string    `gorm:"column:space_type"`
	Deleted        bool      `gorm:"column:deleted"`
}

func (spaceRecord) TableName() string {
	return "space"
}

func toSpaceDomain(r *spaceRecord) (*domainspace.Space, error) {
	st, err := domainspace.NewType(r.SpaceType)
	if err != nil {
		return nil, fmt.Errorf("invalid space type %q: %w", r.SpaceType, err)
	}
	return domainspace.ReconstructSpace(domain.MustParseSpaceID(r.ID), domain.MustParseOrganizationID(r.OrganizationID), domain.MustParseAppUserID(r.OwnerID), r.KeyName, r.Name, st, r.Deleted).
		WithVersion(r.Version), nil
}

// SpaceRepository implements space persistence using GORM.
type SpaceRepository struct {
	db *gorm.DB
}

// NewSpaceRepository returns a new SpaceRepository.
func NewSpaceRepository(db *gorm.DB) *SpaceRepository {
	return &SpaceRepository{db: db}
}

// Save persists a space aggregate. New aggregates (version 1) are inserted;
// loaded aggregates (version > 1) are updated via CAS on the version column.
// The caller is responsible for calling IncrementVersion after a successful Save.
func (r *SpaceRepository) Save(ctx context.Context, space *domainspace.Space) error {
	systemUserID := domain.SystemAppUserID().String()
	record := spaceRecord{
		ID:             space.ID().String(),
		Version:        space.Version(),
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      systemUserID,
		UpdatedBy:      systemUserID,
		OrganizationID: space.OrganizationID().String(),
		OwnerID:        space.OwnerID().String(),
		KeyName:        space.KeyName(),
		Name:           space.Name(),
		SpaceType:      space.SpaceType().Value(),
		Deleted:        space.Deleted(),
	}
	if space.Version() == 1 {
		if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
			return fmt.Errorf("insert space: %w", err)
		}
		return nil
	}

	result := r.db.WithContext(ctx).
		Model(&record).
		Where("id = ? AND version = ?", record.ID, record.Version-1).
		Updates(map[string]any{
			"organization_id": record.OrganizationID,
			"owner_id":        record.OwnerID,
			"key_name":        record.KeyName,
			"name":            record.Name,
			"space_type":      record.SpaceType,
			"deleted":         record.Deleted,
			"version":         record.Version,
		})
	if result.Error != nil {
		return fmt.Errorf("update space: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrSpaceConcurrentModification
	}
	return nil
}

// FindByID looks up a space by its ID.
func (r *SpaceRepository) FindByID(ctx context.Context, id domain.SpaceID) (*domainspace.Space, error) {
	var record spaceRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id.String()).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSpaceNotFound
		}
		return nil, fmt.Errorf("find space by id: %w", err)
	}
	space, err := toSpaceDomain(&record)
	if err != nil {
		return nil, fmt.Errorf("convert space domain: %w", err)
	}
	return space, nil
}

// FindByKeyName looks up a space by organization ID and key name.
func (r *SpaceRepository) FindByKeyName(ctx context.Context, organizationID domain.OrganizationID, keyName string) (*domainspace.Space, error) {
	var record spaceRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND key_name = ?", organizationID.String(), keyName).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSpaceNotFound
		}
		return nil, fmt.Errorf("find space by key name: %w", err)
	}
	space, err := toSpaceDomain(&record)
	if err != nil {
		return nil, fmt.Errorf("convert space domain: %w", err)
	}
	return space, nil
}

// FindByOrganizationID returns all spaces for the given organization.
func (r *SpaceRepository) FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) ([]domainspace.Space, error) {
	var records []spaceRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted = ?", organizationID.String(), false).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find spaces by organization id: %w", err)
	}
	spaces := make([]domainspace.Space, len(records))
	for i := range records {
		space, err := toSpaceDomain(&records[i])
		if err != nil {
			return nil, fmt.Errorf("convert space domain: %w", err)
		}
		spaces[i] = *space
	}
	return spaces, nil
}
