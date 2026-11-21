package gateway

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/mocoarow/cocotola-1.26/cocotola-auth/domain"
)

// findMemberIDs queries records by organization_id and extracts member IDs.
func findMemberIDs[R any, ID any](ctx context.Context, db *gorm.DB, organizationID domain.OrganizationID, extractID func(R) (ID, error), label string) ([]ID, error) {
	var records []R
	if err := db.WithContext(ctx).Where(organizationIDColumn+" = ?", organizationID.String()).Find(&records).Error; err != nil {
		return nil, fmt.Errorf("find %s: %w", label, err)
	}

	ids := make([]ID, len(records))
	for i := range records {
		id, err := extractID(records[i])
		if err != nil {
			return nil, fmt.Errorf("extract id from %s record: %w", label, err)
		}
		ids[i] = id
	}
	return ids, nil
}
