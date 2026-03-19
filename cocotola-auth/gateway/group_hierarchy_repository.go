package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

type groupNGroupRecord struct {
	OrganizationID int       `gorm:"column:organization_id;primaryKey"`
	ParentGroupID  int       `gorm:"column:parent_group_id;primaryKey"`
	ChildGroupID   int       `gorm:"column:child_group_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	CreatedBy      int       `gorm:"column:created_by"`
}

func (groupNGroupRecord) TableName() string {
	return "group_n_group"
}

// GroupHierarchyRepository implements group hierarchy persistence using MySQL via GORM.
type GroupHierarchyRepository struct {
	db *gorm.DB
}

// NewGroupHierarchyRepository returns a new GroupHierarchyRepository.
func NewGroupHierarchyRepository(db *gorm.DB) *GroupHierarchyRepository {
	return &GroupHierarchyRepository{db: db}
}

// FindByOrganizationID returns the group hierarchy for the given organization.
func (r *GroupHierarchyRepository) FindByOrganizationID(ctx context.Context, organizationID int) (*domain.GroupHierarchy, error) {
	var records []groupNGroupRecord
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find group hierarchy by organization id: %w", err)
	}

	edges := make([]domain.HierarchyEdge, len(records))
	for i := range records {
		edges[i] = domain.ReconstructHierarchyEdge(records[i].ParentGroupID, records[i].ChildGroupID)
	}

	hierarchy, err := domain.NewGroupHierarchy(organizationID, edges)
	if err != nil {
		return nil, fmt.Errorf("reconstruct group hierarchy: %w", err)
	}
	return hierarchy, nil
}

// Save persists the group hierarchy by replacing all entries for the organization.
func (r *GroupHierarchyRepository) Save(ctx context.Context, hierarchy *domain.GroupHierarchy) error {
	edges := hierarchy.Edges()
	records := make([]groupNGroupRecord, len(edges))
	for i, edge := range edges {
		records[i] = groupNGroupRecord{
			OrganizationID: hierarchy.OrganizationID(),
			ParentGroupID:  edge.ParentGroupID(),
			ChildGroupID:   edge.ChildGroupID(),
			CreatedAt:      time.Now(),
			CreatedBy:      0,
		}
	}
	return replaceRecords(ctx, r.db, "organization_id = ?", hierarchy.OrganizationID(),
		records, "group hierarchy entries")
}
