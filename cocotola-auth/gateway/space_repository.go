package gateway

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type spaceRecord struct {
	ID             int       `gorm:"column:id;primaryKey"`
	Version        int       `gorm:"column:version;->"`
	CreatedAt      time.Time `gorm:"column:created_at;->"`
	UpdatedAt      time.Time `gorm:"column:updated_at;->"`
	CreatedBy      int       `gorm:"column:created_by;<-:create"`
	UpdatedBy      int       `gorm:"column:updated_by"`
	OrganizationID int       `gorm:"column:organization_id"`
	OwnerID        int       `gorm:"column:owner_id"`
	KeyName        string    `gorm:"column:key_name"`
	Name           string    `gorm:"column:name"`
	SpaceType      string    `gorm:"column:space_type"`
	Deleted        bool      `gorm:"column:deleted"`
}

func (spaceRecord) TableName() string {
	return "space"
}

func toSpaceDomain(r *spaceRecord) (*domain.Space, error) {
	st, err := domain.NewSpaceType(r.SpaceType)
	if err != nil {
		return nil, fmt.Errorf("invalid space type %q: %w", r.SpaceType, err)
	}
	return domain.ReconstructSpace(r.ID, r.OrganizationID, r.OwnerID, r.KeyName, r.Name, st, r.Deleted), nil
}

// SpaceRepository implements space persistence using MySQL via GORM.
type SpaceRepository struct {
	db *gorm.DB
}

// NewSpaceRepository returns a new SpaceRepository.
func NewSpaceRepository(db *gorm.DB) *SpaceRepository {
	return &SpaceRepository{db: db}
}

// Create inserts a new space record and returns the auto-generated ID.
func (r *SpaceRepository) Create(ctx context.Context, organizationID int, ownerID int, keyName string, name string, spaceType string, createdBy int) (int, error) {
	record := spaceRecord{
		ID:             0,
		Version:        0,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      createdBy,
		UpdatedBy:      createdBy,
		OrganizationID: organizationID,
		OwnerID:        ownerID,
		KeyName:        keyName,
		Name:           name,
		SpaceType:      spaceType,
		Deleted:        false,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return 0, fmt.Errorf("create space: %w", err)
	}
	return record.ID, nil
}

// FindByID looks up a space by its ID.
func (r *SpaceRepository) FindByID(ctx context.Context, id int) (*domain.Space, error) {
	var record spaceRecord
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&record).Error; err != nil {
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
func (r *SpaceRepository) FindByKeyName(ctx context.Context, organizationID int, keyName string) (*domain.Space, error) {
	var record spaceRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND key_name = ?", organizationID, keyName).
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
func (r *SpaceRepository) FindByOrganizationID(ctx context.Context, organizationID int) ([]domain.Space, error) {
	var records []spaceRecord
	if err := r.db.WithContext(ctx).
		Where("organization_id = ? AND deleted = 0", organizationID).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find spaces by organization id: %w", err)
	}
	spaces := make([]domain.Space, len(records))
	for i := range records {
		space, err := toSpaceDomain(&records[i])
		if err != nil {
			return nil, fmt.Errorf("convert space domain: %w", err)
		}
		spaces[i] = *space
	}
	return spaces, nil
}

// Save persists a space record (upsert: insert or update).
func (r *SpaceRepository) Save(ctx context.Context, space *domain.Space) error {
	record := spaceRecord{
		ID:             space.ID(),
		Version:        0,
		CreatedAt:      time.Time{},
		UpdatedAt:      time.Time{},
		CreatedBy:      0,
		UpdatedBy:      0,
		OrganizationID: space.OrganizationID(),
		OwnerID:        space.OwnerID(),
		KeyName:        space.KeyName(),
		Name:           space.Name(),
		SpaceType:      space.SpaceType().Value(),
		Deleted:        space.Deleted(),
	}
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("save space: %w", err)
	}
	return nil
}
