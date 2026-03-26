package consultations

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository handles all database operations for the consultations feature.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new consultations repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB exposes the underlying *gorm.DB (used in service transactions).
func (r *Repository) DB() *gorm.DB { return r.db }

// CreateBatch inserts multiple consultation records atomically.
func (r *Repository) CreateBatch(ctx context.Context, consultations []Consultation) error {
	return r.db.WithContext(ctx).Create(&consultations).Error
}

// FindByTokenHash looks up a consultation by the stored SHA-256 token hash.
func (r *Repository) FindByTokenHash(ctx context.Context, hash string) (*Consultation, error) {
	var c Consultation
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", hash).
		First(&c).Error
	return &c, err
}

// FindByApplicationID returns all consultations for an application, ordered by creation time.
func (r *Repository) FindByApplicationID(ctx context.Context, appID string) ([]*Consultation, error) {
	var cs []*Consultation
	err := r.db.WithContext(ctx).
		Where("application_id = ?", appID).
		Order("created_at ASC").
		Find(&cs).Error
	return cs, err
}

// MarkTokenUsed sets token_used_at = now for the given consultation record.
func (r *Repository) MarkTokenUsed(ctx context.Context, consultID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("consultations").
		Where("id = ?", consultID).
		Update("token_used_at", now).Error
}

// SaveResponse stores the member's response (response_type + optional reason).
func (r *Repository) SaveResponse(ctx context.Context, id string, responseType ConsultationResponseType, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("consultations").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"response_type": responseType,
			"reason":        reason,
			"token_used_at": now,
			"updated_at":    now,
		}).Error
}

// CountByType returns how many consultations of a given response type exist for an application.
func (r *Repository) CountByType(ctx context.Context, appID string, responseType ConsultationResponseType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Consultation{}).
		Where("application_id = ? AND response_type = ?", appID, responseType).
		Count(&count).Error
	return count, err
}

// CountTotal returns the total number of consultations for an application.
func (r *Repository) CountTotal(ctx context.Context, appID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Consultation{}).
		Where("application_id = ?", appID).
		Count(&count).Error
	return count, err
}

// CountPending returns consultations with no response yet (token_used_at IS NULL).
func (r *Repository) CountPending(ctx context.Context, appID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Consultation{}).
		Where("application_id = ? AND token_used_at IS NULL", appID).
		Count(&count).Error
	return count, err
}
