package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// --- active group list ---

type activeGroupRecord struct {
	OrganizationID int       `gorm:"column:organization_id;primaryKey"`
	GroupID        int       `gorm:"column:group_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (activeGroupRecord) TableName() string { return "active_group" }

// ActiveGroupListRepository implements active group list persistence using GORM.
type ActiveGroupListRepository struct{ db *gorm.DB }

// NewActiveGroupListRepository returns a new ActiveGroupListRepository.
func NewActiveGroupListRepository(db *gorm.DB) *ActiveGroupListRepository {
	return &ActiveGroupListRepository{db: db}
}

// FindByOrganizationID returns the active group list for the given organization.
func (r *ActiveGroupListRepository) FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveGroupList, error) {
	ids, err := findMemberIDs(ctx, r.db, organizationID,
		func(rec activeGroupRecord) int { return rec.GroupID }, "active groups by organization id")
	if err != nil {
		return nil, err
	}

	list, err := domain.NewActiveGroupList(organizationID, ids)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active group list: %w", err)
	}
	return list, nil
}

// Save persists the active group list by replacing all entries for the organization.
func (r *ActiveGroupListRepository) Save(ctx context.Context, list *domain.ActiveGroupList) error {
	entries := list.Entries()

	records := make([]activeGroupRecord, len(entries))
	for i, groupID := range entries {
		records[i] = activeGroupRecord{
			OrganizationID: list.OrganizationID(),
			GroupID:        groupID,
			CreatedAt:      time.Now(),
		}
	}
	return replaceRecords(ctx, r.db, "organization_id = ?", list.OrganizationID(),
		records, "active group entries")
}

// --- active user list ---

type activeUserRecord struct {
	OrganizationID int       `gorm:"column:organization_id;primaryKey"`
	UserID         int       `gorm:"column:user_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (activeUserRecord) TableName() string { return "active_user" }

// ActiveUserListRepository implements active user list persistence using GORM.
type ActiveUserListRepository struct{ db *gorm.DB }

// NewActiveUserListRepository returns a new ActiveUserListRepository.
func NewActiveUserListRepository(db *gorm.DB) *ActiveUserListRepository {
	return &ActiveUserListRepository{db: db}
}

// FindByOrganizationID returns the active user list for the given organization.
func (r *ActiveUserListRepository) FindByOrganizationID(ctx context.Context, organizationID int) (*domain.ActiveUserList, error) {
	ids, err := findMemberIDs(ctx, r.db, organizationID,
		func(rec activeUserRecord) int { return rec.UserID }, "active users by organization id")
	if err != nil {
		return nil, err
	}

	list, err := domain.NewActiveUserList(organizationID, ids)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active user list: %w", err)
	}
	return list, nil
}

// Save persists the active user list by replacing all entries for the organization.
func (r *ActiveUserListRepository) Save(ctx context.Context, list *domain.ActiveUserList) error {
	entries := list.Entries()

	records := make([]activeUserRecord, len(entries))
	for i, userID := range entries {
		records[i] = activeUserRecord{
			OrganizationID: list.OrganizationID(),
			UserID:         userID,
			CreatedAt:      time.Now(),
		}
	}
	return replaceRecords(ctx, r.db, "organization_id = ?", list.OrganizationID(),
		records, "active user entries")
}
