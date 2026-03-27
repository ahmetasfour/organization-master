package webpublish

import (
	"context"
	"errors"
	"fmt"
	"time"

	"membership-system/api/internal/features/applications"
	"membership-system/api/internal/features/logs"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrApplicationNotAccepted = errors.New("consent can only be recorded for accepted applications")
	ErrConsentAlreadyRecorded = errors.New("consent already recorded for this application")
	ErrApplicationNotFound    = errors.New("application not found")
)

// Service handles web publish business logic
type Service struct {
	repo       *Repository
	logService *logs.Service
	db         *gorm.DB
}

// NewService creates a new web publish service
func NewService(repo *Repository, logService *logs.Service, db *gorm.DB) *Service {
	return &Service{
		repo:       repo,
		logService: logService,
		db:         db,
	}
}

// RecordConsent records the web publish consent for an accepted application
func (s *Service) RecordConsent(ctx context.Context, applicationID string, req *RecordConsentRequest, adminID string) (*ConsentResponse, error) {
	// Load application
	app, err := s.repo.GetApplicationByID(ctx, applicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, fmt.Errorf("service.RecordConsent: failed to load application: %w", err)
	}

	// Assert application status is kabul
	if app.Status != applications.StatusKabul {
		return nil, ErrApplicationNotAccepted
	}

	// Check if consent already recorded
	exists, err := s.repo.ConsentExists(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("service.RecordConsent: failed to check existing consent: %w", err)
	}
	if exists {
		return nil, ErrConsentAlreadyRecorded
	}

	// Execute in transaction
	var consent *WebPublishConsent
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &Repository{db: tx}

		// Create consent record
		consent = &WebPublishConsent{
			ID:            uuid.New().String(),
			ApplicationID: applicationID,
			Consented:     req.Consented,
			RecordedBy:    adminID,
			CreatedAt:     time.Now(),
		}

		if err := txRepo.RecordConsent(ctx, consent); err != nil {
			return fmt.Errorf("failed to create consent record: %w", err)
		}

		// Update application publish status
		if err := txRepo.UpdateApplicationPublishStatus(ctx, applicationID, req.Consented); err != nil {
			return fmt.Errorf("failed to update application publish status: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("service.RecordConsent: transaction failed: %w", err)
	}

	// Log the action
	metadata := map[string]interface{}{
		"consented":      req.Consented,
		"application_id": applicationID,
		"admin_id":       adminID,
	}
	if err := s.logService.Create(ctx, adminID, "admin", "publish.consent_recorded", "application", applicationID, metadata, ""); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to log consent recording: %v\n", err)
	}

	return &ConsentResponse{
		ApplicationID: applicationID,
		Consented:     req.Consented,
		IsPublished:   req.Consented,
		RecordedAt:    consent.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetPublishedMembers returns the list of published members sorted alphabetically
func (s *Service) GetPublishedMembers(ctx context.Context) ([]*MemberListItem, error) {
	apps, err := s.repo.FindPublishedMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("service.GetPublishedMembers: %w", err)
	}

	members := make([]*MemberListItem, len(apps))
	for i, app := range apps {
		members[i] = &MemberListItem{
			FullName:       app.ApplicantName,
			MembershipType: string(app.MembershipType),
			AcceptedAt:     app.UpdatedAt.Format("2006-01-02"),
		}
	}

	return members, nil
}

// GetConsentStatus returns the consent status for an application
func (s *Service) GetConsentStatus(ctx context.Context, applicationID string) (*ConsentResponse, error) {
	consent, err := s.repo.GetConsentByApplicationID(ctx, applicationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No consent recorded yet
		}
		return nil, fmt.Errorf("service.GetConsentStatus: %w", err)
	}

	return &ConsentResponse{
		ApplicationID: applicationID,
		Consented:     consent.Consented,
		IsPublished:   consent.Consented,
		RecordedAt:    consent.CreatedAt.Format(time.RFC3339),
	}, nil
}
