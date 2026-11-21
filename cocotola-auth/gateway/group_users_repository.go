package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

type userNGroupRecord struct {
	GroupID   string    `gorm:"column:group_id;primaryKey"`
	UserID    string    `gorm:"column:user_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
	CreatedBy string    `gorm:"column:created_by"`
}

func (userNGroupRecord) TableName() string {
	return "user_n_group"
}

const userNGroupRootTable = "user_n_group_root"

// GroupUsersRepository implements group-user association persistence using GORM.
type GroupUsersRepository struct {
	db *gorm.DB
}

// NewGroupUsersRepository returns a new GroupUsersRepository.
func NewGroupUsersRepository(db *gorm.DB) *GroupUsersRepository {
	return &GroupUsersRepository{db: db}
}

// FindByGroupID returns the group users aggregate for the given group.
func (r *GroupUsersRepository) FindByGroupID(ctx context.Context, groupID domain.GroupID) (*domaingroup.Users, error) {
	version, err := findRootVersion(ctx, r.db, userNGroupRootTable, groupIDColumn, groupID.String())
	if err != nil {
		return nil, err
	}

	var records []userNGroupRecord
	if err := r.db.WithContext(ctx).Where(groupIDColumn+" = ?", groupID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find group users by group id: %w", err)
	}

	userIDs := make([]domain.AppUserID, len(records))
	for i := range records {
		uid, err := domain.ParseAppUserID(records[i].UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user id %q in db: %w", records[i].UserID, err)
		}
		userIDs[i] = uid
	}

	gu, err := domaingroup.NewUsers(groupID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct group users: %w", err)
	}
	gu.SetVersion(version)
	return gu, nil
}

// Save persists the group users aggregate with an optimistic lock on its root row.
func (r *GroupUsersRepository) Save(ctx context.Context, gu *domaingroup.Users) error {
	groupIDStr := gu.GroupID().String()
	userIDs := gu.UserIDs()
	systemUserID := domain.SystemAppUserID().String()
	records := make([]userNGroupRecord, len(userIDs))
	for i, userID := range userIDs {
		records[i] = userNGroupRecord{
			GroupID:   groupIDStr,
			UserID:    userID.String(),
			CreatedAt: time.Now(),
			CreatedBy: systemUserID,
		}
	}
	root := newGroupRootRecord(groupIDStr, gu)
	if err := saveVersionedMembers(ctx, r.db, userNGroupRootTable, gu, root, groupIDColumn, groupIDStr, records); err != nil {
		return fmt.Errorf("save users of group %s: %w", groupIDStr, err)
	}
	return nil
}
