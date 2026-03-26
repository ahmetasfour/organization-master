package logs

import (
	"context"

	"gorm.io/gorm"
)

// Repository handles database operations for logs
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new logs repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new log entry
func (r *Repository) Create(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByEntityID finds all logs for a specific entity
func (r *Repository) FindByEntityID(ctx context.Context, entityType, entityID string) ([]Log, error) {
	var logs []Log
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// FindByAction finds all logs for a specific action
func (r *Repository) FindByAction(ctx context.Context, action string) ([]Log, error) {
	var logs []Log
	err := r.db.WithContext(ctx).
		Where("action = ?", action).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}
