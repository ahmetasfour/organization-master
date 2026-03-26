package reputation

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository handles all DB operations for the reputation feature.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new reputation repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB exposes the underlying *gorm.DB for use in service-level transactions.
func (r *Repository) DB() *gorm.DB { return r.db }

// CreateBatch inserts all reputation contacts for an application in one call.
func (r *Repository) CreateBatch(ctx context.Context, contacts []ReputationContact) error {
	return r.db.WithContext(ctx).Create(&contacts).Error
}

// FindByApplicationID returns all contacts for an application, ordered by creation time.
func (r *Repository) FindByApplicationID(ctx context.Context, appID string) ([]*ReputationContact, error) {
	var contacts []*ReputationContact
	err := r.db.WithContext(ctx).
		Where("application_id = ?", appID).
		Order("created_at ASC").
		Find(&contacts).Error
	return contacts, err
}

// FindByTokenHash looks up a contact by the stored SHA-256 token hash.
func (r *Repository) FindByTokenHash(ctx context.Context, hash string) (*ReputationContact, error) {
	var c ReputationContact
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&c).Error
	return &c, err
}

// MarkTokenUsed atomically sets token_used_at = now for the given contact.
func (r *Repository) MarkTokenUsed(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&ReputationContact{}).
		Where("id = ?", id).
		Update("token_used_at", now).Error
}

// SaveResponse persists the contact's response fields in the DB.
func (r *Repository) SaveResponse(
	ctx context.Context,
	id string,
	responseType ReputationResponseType,
	reason string,
	ip string,
) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&ReputationContact{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"response_type": responseType,
			"reason":        reason,
			"responded_at":  now,
			"responded_ip":  ip,
		}).Error
}

// CountByApplicationAndType counts contacts with a given response type for an application.
func (r *Repository) CountByApplicationAndType(
	ctx context.Context,
	appID string,
	responseType ReputationResponseType,
) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReputationContact{}).
		Where("application_id = ? AND response_type = ?", appID, responseType).
		Count(&count).Error
	return count, err
}

// CountTotal returns the total number of contacts for an application.
func (r *Repository) CountTotal(ctx context.Context, appID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReputationContact{}).
		Where("application_id = ?", appID).
		Count(&count).Error
	return count, err
}

// CountResponded returns the number of contacts that have already responded.
func (r *Repository) CountResponded(ctx context.Context, appID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReputationContact{}).
		Where("application_id = ? AND response_type IS NOT NULL", appID).
		Count(&count).Error
	return count, err
}
