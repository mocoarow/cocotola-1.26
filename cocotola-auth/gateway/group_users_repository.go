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
	GroupID   int       `gorm:"column:group_id;primaryKey"`
	UserID    string    `gorm:"column:user_id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
	CreatedBy string    `gorm:"column:created_by"`
}

func (userNGroupRecord) TableName() string {
	return "user_n_group"
}

// GroupUsersRepository implements group-user association persistence using GORM.
type GroupUsersRepository struct {
	db *gorm.DB
}

// NewGroupUsersRepository returns a new GroupUsersRepository.
func NewGroupUsersRepository(db *gorm.DB) *GroupUsersRepository {
	return &GroupUsersRepository{db: db}
}

// FindByGroupID returns the group users aggregate for the given group.
func (r *GroupUsersRepository) FindByGroupID(ctx context.Context, groupID int) (*domaingroup.Users, error) {
	var records []userNGroupRecord
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find group users by group id: %w", err)
	}

	userIDs := make([]domain.AppUserID, len(records))
	for i := range records {
		userIDs[i] = domain.MustParseAppUserID(records[i].UserID)
	}

	gu, err := domaingroup.NewUsers(groupID, userIDs)
	if err != nil {
		return nil, fmt.Errorf("reconstruct group users: %w", err)
	}
	return gu, nil
}

// Save persists the group users aggregate by replacing all entries for the group.
func (r *GroupUsersRepository) Save(ctx context.Context, gu *domaingroup.Users) error {
	userIDs := gu.UserIDs()
	systemUserID := domain.SystemAppUserID().String()
	records := make([]userNGroupRecord, len(userIDs))
	for i, userID := range userIDs {
		records[i] = userNGroupRecord{
			GroupID:   gu.GroupID(),
			UserID:    userID.String(),
			CreatedAt: time.Now(),
			CreatedBy: systemUserID,
		}
	}
	return replaceRecords(ctx, r.db, "group_id = ?", gu.GroupID(),
		records, "group user entries")
}
