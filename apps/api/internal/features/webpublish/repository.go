package webpublish

import (
	"context"

	"membership-system/api/internal/features/applications"

	"gorm.io/gorm"
)

// Repository handles database operations for web publish consent
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new web publish repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// RecordConsent creates a new consent record
func (r *Repository) RecordConsent(ctx context.Context, consent *WebPublishConsent) error {
	return r.db.WithContext(ctx).Create(consent).Error
}

// ConsentExists checks if consent has already been recorded for an application
func (r *Repository) ConsentExists(ctx context.Context, applicationID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&WebPublishConsent{}).
		Where("application_id = ?", applicationID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetConsentByApplicationID retrieves consent for an application
func (r *Repository) GetConsentByApplicationID(ctx context.Context, applicationID string) (*WebPublishConsent, error) {
	var consent WebPublishConsent
	err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		First(&consent).Error
	if err != nil {
		return nil, err
	}
	return &consent, nil
}

// FindPublishedMembers returns all published members ordered alphabetically
func (r *Repository) FindPublishedMembers(ctx context.Context) ([]*applications.Application, error) {
	var apps []*applications.Application
	err := r.db.WithContext(ctx).
		Where("is_published = ? AND status = ?", true, applications.StatusKabul).
		Order("applicant_name ASC").
		Find(&apps).Error
	if err != nil {
		return nil, err
	}
	return apps, nil
}

// UpdateApplicationPublishStatus updates the application's publish consent and status
func (r *Repository) UpdateApplicationPublishStatus(ctx context.Context, applicationID string, consented bool) error {
	updates := map[string]interface{}{
		"web_publish_consent": consented,
		"is_published":        consented,
	}
	return r.db.WithContext(ctx).
		Model(&applications.Application{}).
		Where("id = ?", applicationID).
		Updates(updates).Error
}

// GetApplicationByID retrieves an application by ID
func (r *Repository) GetApplicationByID(ctx context.Context, id string) (*applications.Application, error) {
	var app applications.Application
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&app).Error
	if err != nil {
		return nil, err
	}
	return &app, nil
}
