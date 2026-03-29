package gateway

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// HealthRepository checks database connectivity.
type HealthRepository struct {
	db *gorm.DB
}

// NewHealthRepository returns a new HealthRepository.
func NewHealthRepository(db *gorm.DB) *HealthRepository {
	return &HealthRepository{db: db}
}

// Check verifies database connectivity by executing SELECT 1.
func (r *HealthRepository) Check(ctx context.Context) error {
	var result int
	if err := r.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return fmt.Errorf("health check query: %w", err)
	}

	return nil
}
