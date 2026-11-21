package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	libversioned "github.com/mocoarow/cocotola-1.26/cocotola-lib/domain/versioned"
)

func saveActiveList[R schema.Tabler](ctx context.Context, db *gorm.DB, rootTable string, list libversioned.Entity, organizationID string, records []R) error {
	root := newOrganizationRootRecord(organizationID, list)
	if err := saveVersionedMembers(ctx, db, rootTable, list, root, organizationIDColumn, organizationID, records); err != nil {
		return fmt.Errorf("save active list of organization %s: %w", organizationID, err)
	}
	return nil
}

// --- active group list ---

type activeGroupRecord struct {
	OrganizationID string    `gorm:"column:organization_id;primaryKey"`
	GroupID        string    `gorm:"column:group_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (activeGroupRecord) TableName() string { return "active_group" }

const activeGroupListTable = "active_group_list"

// ActiveGroupListRepository implements active group list persistence using GORM.
type ActiveGroupListRepository struct{ db *gorm.DB }

// NewActiveGroupListRepository returns a new ActiveGroupListRepository.
func NewActiveGroupListRepository(db *gorm.DB) *ActiveGroupListRepository {
	return &ActiveGroupListRepository{db: db}
}

// FindByOrganizationID returns the active group list for the given organization.
func (r *ActiveGroupListRepository) FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*domain.ActiveGroupList, error) {
	version, err := findRootVersion(ctx, r.db, activeGroupListTable, organizationIDColumn, organizationID.String())
	if err != nil {
		return nil, err
	}
	ids, err := findMemberIDs(ctx, r.db, organizationID,
		func(rec activeGroupRecord) (domain.GroupID, error) { return domain.ParseGroupID(rec.GroupID) }, "active groups by organization id")
	if err != nil {
		return nil, fmt.Errorf("find active group ids: %w", err)
	}

	list, err := domain.NewActiveGroupList(organizationID, ids)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active group list: %w", err)
	}
	list.SetVersion(version)
	return list, nil
}

// Save persists the active group list with an optimistic lock on its root row.
func (r *ActiveGroupListRepository) Save(ctx context.Context, list *domain.ActiveGroupList) error {
	orgIDStr := list.OrganizationID().String()
	entries := list.Entries()
	records := make([]activeGroupRecord, len(entries))
	for i, groupID := range entries {
		records[i] = activeGroupRecord{OrganizationID: orgIDStr, GroupID: groupID.String(), CreatedAt: time.Now()}
	}
	return saveActiveList(ctx, r.db, activeGroupListTable, list, orgIDStr, records)
}

// --- active user list ---

type activeUserRecord struct {
	OrganizationID string    `gorm:"column:organization_id;primaryKey"`
	UserID         string    `gorm:"column:user_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (activeUserRecord) TableName() string { return "active_user" }

const activeUserListTable = "active_user_list"

// ActiveUserListRepository implements active user list persistence using GORM.
type ActiveUserListRepository struct{ db *gorm.DB }

// NewActiveUserListRepository returns a new ActiveUserListRepository.
func NewActiveUserListRepository(db *gorm.DB) *ActiveUserListRepository {
	return &ActiveUserListRepository{db: db}
}

// FindByOrganizationID returns the active user list for the given organization.
func (r *ActiveUserListRepository) FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*domain.ActiveUserList, error) {
	version, err := findRootVersion(ctx, r.db, activeUserListTable, organizationIDColumn, organizationID.String())
	if err != nil {
		return nil, err
	}
	ids, err := findMemberIDs(ctx, r.db, organizationID,
		func(rec activeUserRecord) (domain.AppUserID, error) { return domain.ParseAppUserID(rec.UserID) }, "active users by organization id")
	if err != nil {
		return nil, fmt.Errorf("find active user ids: %w", err)
	}

	list, err := domain.NewActiveUserList(organizationID, ids)
	if err != nil {
		return nil, fmt.Errorf("reconstruct active user list: %w", err)
	}
	list.SetVersion(version)
	return list, nil
}

// Save persists the active user list with an optimistic lock on its root row.
func (r *ActiveUserListRepository) Save(ctx context.Context, list *domain.ActiveUserList) error {
	orgIDStr := list.OrganizationID().String()
	entries := list.Entries()
	records := make([]activeUserRecord, len(entries))
	for i, userID := range entries {
		records[i] = activeUserRecord{OrganizationID: orgIDStr, UserID: userID.String(), CreatedAt: time.Now()}
	}
	return saveActiveList(ctx, r.db, activeUserListTable, list, orgIDStr, records)
}
