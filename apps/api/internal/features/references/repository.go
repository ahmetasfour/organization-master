package references

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// Repository handles all database operations for the references feature.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new references repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB exposes the underlying *gorm.DB for use in transactions.
func (r *Repository) DB() *gorm.DB { return r.db }

// CreateBatch inserts multiple reference records atomically.
func (r *Repository) CreateBatch(ctx context.Context, refs []Reference) error {
	return r.db.WithContext(ctx).Create(&refs).Error
}

// FindByTokenHash looks up a reference (with its response) by the stored SHA-256 token hash.
func (r *Repository) FindByTokenHash(ctx context.Context, hash string) (*Reference, error) {
	var ref Reference
	err := r.db.WithContext(ctx).
		Preload("Response").
		Where("token_hash = ?", hash).
		First(&ref).Error
	return &ref, err
}

// FindByID finds a reference by its primary key.
func (r *Repository) FindByID(ctx context.Context, id string) (*Reference, error) {
	var ref Reference
	err := r.db.WithContext(ctx).
		Preload("Response").
		Where("id = ?", id).
		First(&ref).Error
	return &ref, err
}

// FindByApplicationID returns all references for an application (with responses).
func (r *Repository) FindByApplicationID(ctx context.Context, appID string) ([]*Reference, error) {
	var refs []*Reference
	err := r.db.WithContext(ctx).
		Preload("Response").
		Where("application_id = ?", appID).
		Order("round ASC, created_at ASC").
		Find(&refs).Error
	return refs, err
}

// MarkTokenUsed sets token_used_at = now for the given reference.
func (r *Repository) MarkTokenUsed(ctx context.Context, refID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Table("references").
		Where("id = ?", refID).
		Update("token_used_at", now).Error
}

// CreateResponse inserts a ReferenceResponse record.
func (r *Repository) CreateResponse(ctx context.Context, response *ReferenceResponse) error {
	return r.db.WithContext(ctx).Create(response).Error
}

// UpdateToken regenerates the token hash and expiry for an existing reference
// (used when resending or creating a replacement).
func (r *Repository) UpdateToken(ctx context.Context, refID, tokenHash string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).
		Table("references").
		Where("id = ?", refID).
		Updates(map[string]interface{}{
			"token_hash":       tokenHash,
			"token_expires_at": expiresAt,
			"token_used_at":    nil,
		}).Error
}

// CountResponsesByType returns how many reference responses of a given type exist for an application.
func (r *Repository) CountResponsesByType(ctx context.Context, appID string, responseType ResponseType) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ReferenceResponse{}).
		Joins("JOIN `references` ON `references`.id = reference_responses.reference_id").
		Where("`references`.application_id = ? AND reference_responses.response_type = ?", appID, responseType).
		Count(&count).Error
	return count, err
}

// CountPending returns the number of references that have not yet been responded to
// (token_used_at IS NULL).
func (r *Repository) CountPending(ctx context.Context, appID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Reference{}).
		Where("application_id = ? AND token_used_at IS NULL", appID).
		Count(&count).Error
	return count, err
}

// IsComplete checks whether the application has enough positive references
// and no pending (unanswered) references remaining.
// Completion criteria: positive_count >= 3 AND pending_count == 0.
func (r *Repository) IsComplete(ctx context.Context, appID string) (bool, error) {
	pending, err := r.CountPending(ctx, appID)
	if err != nil {
		return false, err
	}
	if pending > 0 {
		return false, nil
	}

	positive, err := r.CountResponsesByType(ctx, appID, ResponsePositive)
	if err != nil {
		return false, err
	}
	return positive >= 3, nil
}

// CreateReplacement creates a new reference slot for an unknown response.
// The caller is responsible for generating and sending a token.
func (r *Repository) CreateReplacement(ctx context.Context, appID, refereeName, refereeEmail string, round int) (*Reference, error) {
	ref := &Reference{
		ApplicationID: appID,
		RefereeName:   refereeName,
		RefereeEmail:  refereeEmail,
		IsReplacement: true,
		Round:         round,
		// TokenHash / TokenExpiresAt will be set by the caller before persisting.
	}
	if err := r.db.WithContext(ctx).Create(ref).Error; err != nil {
		return nil, err
	}
	return ref, nil
}

// FindReplacementByToken finds a replacement reference by its token hash.
// Returns the reference with preloaded application data (applicant_name, membership_type).
func (r *Repository) FindReplacementByToken(ctx context.Context, tokenHash string) (*Reference, *AppContext, error) {
	var ref Reference
	if err := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_replacement = true", tokenHash).
		First(&ref).Error; err != nil {
		return nil, nil, err
	}

	// Load application context
	var app struct {
		ID             string `gorm:"column:id"`
		ApplicantName  string `gorm:"column:applicant_name"`
		ApplicantEmail string `gorm:"column:applicant_email"`
		MembershipType string `gorm:"column:membership_type"`
	}
	if err := r.db.WithContext(ctx).Table("applications").
		Where("id = ?", ref.ApplicationID).
		First(&app).Error; err != nil {
		return nil, nil, err
	}

	appCtx := &AppContext{
		ID:             app.ID,
		ApplicantName:  app.ApplicantName,
		ApplicantEmail: app.ApplicantEmail,
		MembershipType: app.MembershipType,
	}

	return &ref, appCtx, nil
}

// UpdateRefereeInfo updates the referee name and email for a replacement reference.
func (r *Repository) UpdateRefereeInfo(ctx context.Context, refID, refereeName, refereeEmail string) error {
	return r.db.WithContext(ctx).
		Model(&Reference{}).
		Where("id = ?", refID).
		Updates(map[string]interface{}{
			"referee_name":  refereeName,
			"referee_email": refereeEmail,
		}).Error
}
