package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type userNGroupRecord struct {
	GroupID   int       `gorm:"column:group_id;primaryKey"`
	UserID    int       `gorm:"column:user_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
	CreatedBy int       `gorm:"column:created_by"`
}

func (userNGroupRecord) TableName() string {
	return "user_n_group"
}

// GroupUsersRepository implements group-user association persistence using MySQL via GORM.
type GroupUsersRepository struct {
	db *gorm.DB
}

// NewGroupUsersRepository returns a new GroupUsersRepository.
func NewGroupUsersRepository(db *gorm.DB) *GroupUsersRepository {
	return &GroupUsersRepository{db: db}
}

// FindByGroupID returns the group users aggregate for the given group.
func (r *GroupUsersRepository) FindByGroupID(ctx context.Context, groupID int) (*domain.GroupUsers, error) {
	var records []userNGroupRecord
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find group users by group id: %w", err)
	}

	userIDs := make([]int, len(records))
	for i := range records {
		userIDs[i] = records[i].UserID
	}

	gu, err := domain.NewGroupUsers(groupID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct group users: %w", err)
	}
	return gu, nil
}

// Save persists the group users aggregate by replacing all entries for the group.
func (r *GroupUsersRepository) Save(ctx context.Context, gu *domain.GroupUsers) error {
	userIDs := gu.UserIDs()
	records := make([]userNGroupRecord, len(userIDs))
	for i, userID := range userIDs {
		records[i] = userNGroupRecord{
			GroupID:   gu.GroupID(),
			UserID:    userID,
			CreatedAt: time.Now(),
			CreatedBy: 0,
		}
	}
	return replaceRecords(ctx, r.db, "group_id = ?", gu.GroupID(),
		&userNGroupRecord{GroupID: 0, UserID: 0, CreatedAt: time.Time{}, CreatedBy: 0},
		records, "group user entries")
}
