package gateway

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
	domaingroup "github.com/mocoarow/cocotola-1.26/cocotola-auth/domain/group"
)

type groupNGroupRecord struct {
	OrganizationID string    `gorm:"column:organization_id;primaryKey"`
	ParentGroupID  string    `gorm:"column:parent_group_id;primaryKey"`
	ChildGroupID   string    `gorm:"column:child_group_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	CreatedBy      string    `gorm:"column:created_by"`
}

func (groupNGroupRecord) TableName() string {
	return "group_n_group"
}

// GroupHierarchyRepository implements group hierarchy persistence using GORM.
type GroupHierarchyRepository struct {
	db *gorm.DB
}

// NewGroupHierarchyRepository returns a new GroupHierarchyRepository.
func NewGroupHierarchyRepository(db *gorm.DB) *GroupHierarchyRepository {
	return &GroupHierarchyRepository{db: db}
}

// FindByOrganizationID returns the group hierarchy for the given organization.
func (r *GroupHierarchyRepository) FindByOrganizationID(ctx context.Context, organizationID domain.OrganizationID) (*domaingroup.Hierarchy, error) {
	var records []groupNGroupRecord
	if err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find group hierarchy by organization id: %w", err)
	}

	edges := make([]domaingroup.HierarchyEdge, len(records))
	for i := range records {
		edges[i] = domaingroup.ReconstructHierarchyEdge(domain.MustParseGroupID(records[i].ParentGroupID), domain.MustParseGroupID(records[i].ChildGroupID))
	}

	hierarchy, err := domaingroup.NewHierarchy(organizationID, edges)
	if err != nil {
		return nil, fmt.Errorf("reconstruct group hierarchy: %w", err)
	}
	return hierarchy, nil
}

// Save persists the group hierarchy by replacing all entries for the organization.
func (r *GroupHierarchyRepository) Save(ctx context.Context, hierarchy *domaingroup.Hierarchy) error {
	edges := hierarchy.Edges()
	orgIDStr := hierarchy.OrganizationID().String()
	systemUserID := domain.SystemAppUserID().String()
	records := make([]groupNGroupRecord, len(edges))
	for i, edge := range edges {
		records[i] = groupNGroupRecord{
			OrganizationID: orgIDStr,
			ParentGroupID:  edge.ParentGroupID().String(),
			ChildGroupID:   edge.ChildGroupID().String(),
			CreatedAt:      time.Now(),
			CreatedBy:      systemUserID,
		}
	}
	return replaceRecords(ctx, r.db, "organization_id = ?", orgIDStr,
		records, "group hierarchy entries")
}
