package consultations

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/features/notifications"
	"membership-system/api/internal/shared"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// appRow is a minimal struct for loading application data without importing the
// applications package (avoids circular dependency).
type appRow struct {
	ID             string `gorm:"column:id"`
	MembershipType string `gorm:"column:membership_type"`
	Status         string `gorm:"column:status"`
	ApplicantName  string `gorm:"column:applicant_name"`
	ApplicantEmail string `gorm:"column:applicant_email"`
	LinkedInURL    string `gorm:"column:linkedin_url"`
}

func (appRow) TableName() string { return "applications" }

// consultTypes are the only membership types that use the consultation flow.
var consultTypes = map[string]bool{
	"profesyonel": true,
	"öğrenci":     true,
}

// Service contains the business logic for the consultation system.
type Service struct {
	repo      *Repository
	authRepo  *auth.Repository
	logRepo   *logs.Repository
	notifySvc *notifications.Service
	db        *gorm.DB
}

// NewService creates a new consultation service.
func NewService(
	repo *Repository,
	authRepo *auth.Repository,
	logRepo *logs.Repository,
	notifySvc *notifications.Service,
	db *gorm.DB,
) *Service {
	return &Service{
		repo:      repo,
		authRepo:  authRepo,
		logRepo:   logRepo,
		notifySvc: notifySvc,
		db:        db,
	}
}

// ─── AddConsultees ────────────────────────────────────────────────────────────

// AddConsultees validates the request, creates consultation records, sends
// tokenized emails, and advances the application status → danışma_sürecinde.
func (s *Service) AddConsultees(ctx context.Context, appID string, req *AddConsultationsRequest, koordinatorID string) error {
	if len(req.Consultees) < 2 {
		return fmt.Errorf("consultations: minimum 2 consultees required")
	}

	// Load application
	var app appRow
	if err := s.db.WithContext(ctx).First(&app, "id = ?", appID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrNotFound
		}
		return fmt.Errorf("consultations: load application: %w", err)
	}

	// Validate membership type
	if !consultTypes[app.MembershipType] {
		return fmt.Errorf("consultations: consultation flow only applies to profesyonel/öğrenci applications")
	}

	// Validate status
	if app.Status != "başvuru_alındı" {
		return fmt.Errorf("consultations: application must be in başvuru_alındı status, got: %s", app.Status)
	}

	// RedGuard: must not be terminated
	redGuard := shared.NewRedGuard(s.db)
	if err := redGuard.AssertNotTerminated(ctx, appID); err != nil {
		return err
	}

	// Build consultation records
	consultations := make([]Consultation, 0, len(req.Consultees))
	type pendingEmail struct {
		consultation Consultation
		rawToken     string
	}
	pending := make([]pendingEmail, 0, len(req.Consultees))

	for _, input := range req.Consultees {
		// Validate member exists and is active
		user, err := s.authRepo.FindByID(ctx, input.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("consultations: user %s not found", input.UserID)
			}
			return fmt.Errorf("consultations: lookup user: %w", err)
		}
		if !user.IsActive {
			return fmt.Errorf("consultations: user %s is inactive", input.UserID)
		}

		tok := shared.GenerateToken()

		c := Consultation{
			ApplicationID:  appID,
			MemberUserID:   user.ID,
			MemberName:     user.FullName,
			MemberEmail:    user.Email,
			TokenHash:      tok.HashedToken,
			TokenExpiresAt: tok.ExpiresAt,
		}
		consultations = append(consultations, c)
		pending = append(pending, pendingEmail{consultation: c, rawToken: tok.RawToken})
	}

	// Persist all records
	if err := s.repo.CreateBatch(ctx, consultations); err != nil {
		return fmt.Errorf("consultations: create batch: %w", err)
	}

	// Send emails (non-fatal per member — log failures and continue)
	for _, p := range pending {
		if err := s.notifySvc.SendConsultationRequest(
			ctx,
			p.consultation.MemberEmail,
			p.consultation.MemberName,
			p.rawToken,
			app.ApplicantName,
			app.MembershipType,
			app.LinkedInURL,
			p.consultation.TokenExpiresAt,
		); err != nil {
			_ = s.writeLog(ctx, "consult.email_failed", appID, "application", map[string]interface{}{
				"member_email": p.consultation.MemberEmail,
				"error":        err.Error(),
			})
		} else {
			_ = s.writeLog(ctx, "consult.sent", appID, "application", map[string]interface{}{
				"member_id":    p.consultation.MemberUserID,
				"member_email": p.consultation.MemberEmail,
			})
		}
	}

	// Advance application status → danışma_sürecinde
	if err := s.db.WithContext(ctx).
		Exec("UPDATE applications SET status = ?, updated_at = ? WHERE id = ?",
			"danışma_sürecinde", time.Now(), appID).Error; err != nil {
		return fmt.Errorf("consultations: advance status: %w", err)
	}

	_ = s.writeLog(ctx, "status.change", appID, "application", map[string]interface{}{
		"from":   "başvuru_alındı",
		"to":     "danışma_sürecinde",
		"reason": "consultations added",
	})

	return nil
}

// ─── GetFormData ──────────────────────────────────────────────────────────────

// GetFormData retrieves public form data for a raw consultation token.
// Does NOT consume the token — only verifies expiry and used status.
func (s *Service) GetFormData(ctx context.Context, rawToken string) (*ConsultationFormData, error) {
	hash := shared.HashToken(rawToken)

	c, err := s.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("consultations: find by token: %w", err)
	}

	if c.IsTokenExpired() {
		return nil, shared.ErrTokenExpired
	}
	if c.IsTokenUsed() {
		return nil, shared.ErrTokenUsed
	}

	// Load application for display context
	var app appRow
	if err := s.db.WithContext(ctx).First(&app, "id = ?", c.ApplicationID).Error; err != nil {
		return nil, fmt.Errorf("consultations: load application: %w", err)
	}

	return &ConsultationFormData{
		ApplicantName:  app.ApplicantName,
		MembershipType: app.MembershipType,
		MemberName:     c.MemberName,
		ExpiresAt:      notifications.FormatTime(c.TokenExpiresAt),
	}, nil
}

// ─── SubmitResponse ───────────────────────────────────────────────────────────

// SubmitResponse validates and consumes a consultation token, saves the
// member's response, and applies the appropriate business logic:
//   - negative → RedGuard.Terminate (application permanently rejected)
//   - positive → if all consultations positive → advance to gündemde
func (s *Service) SubmitResponse(ctx context.Context, rawToken string, req *ConsultationResponseRequest, ipAddress string) error {
	if req.ResponseType == string(ConsultResponseNegative) && len(req.Reason) < 30 {
		return fmt.Errorf("reason: minimum 30 characters required for negative response")
	}

	hash := shared.HashToken(rawToken)
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Validate and atomically consume token
		if err := shared.ValidateAndConsumeToken(ctx, tx, "consultations", hash, now); err != nil {
			return err
		}

		// 2. Reload consultation within transaction
		var c Consultation
		if err := tx.Where("token_hash = ?", hash).First(&c).Error; err != nil {
			return fmt.Errorf("consultations: reload: %w", err)
		}

		// 3. Save response
		respType := ConsultationResponseType(req.ResponseType)
		if err := tx.Table("consultations").
			Where("id = ?", c.ID).
			Updates(map[string]interface{}{
				"response_type": respType,
				"reason":        req.Reason,
				"updated_at":    now,
			}).Error; err != nil {
			return fmt.Errorf("consultations: save response: %w", err)
		}

		// 4. Audit log
		_ = writeLogTx(ctx, tx, "consult.responded", c.ApplicationID, "application", map[string]interface{}{
			"response_type": req.ResponseType,
			"ip":            ipAddress,
			"consult_id":    c.ID,
		})

		// 5. Business logic
		switch respType {
		case ConsultResponseNegative:
			redGuard := shared.NewRedGuard(tx)
			if err := redGuard.Terminate(ctx, c.ApplicationID, req.Reason, "system", "consulted_member"); err != nil {
				if errors.Is(err, shared.ErrApplicationTerminated) {
					return nil // idempotent
				}
				return fmt.Errorf("consultations: terminate: %w", err)
			}

			// Send rejection email (reason NOT included per spec)
			var appInfo struct {
				ApplicantName  string `gorm:"column:applicant_name"`
				ApplicantEmail string `gorm:"column:applicant_email"`
			}
			if err := tx.Table("applications").
				Select("applicant_name", "applicant_email").
				Where("id = ?", c.ApplicationID).
				First(&appInfo).Error; err == nil {
				_ = s.notifySvc.SendRejected(ctx, c.ApplicationID, appInfo.ApplicantEmail, appInfo.ApplicantName)
			}

		case ConsultResponsePositive:
			// Check if ALL consultations for this application are positive (none pending)
			var pending int64
			tx.Model(&Consultation{}).
				Where("application_id = ? AND token_used_at IS NULL", c.ApplicationID).
				Count(&pending)

			if pending == 0 {
				// Ensure no negative responses exist
				var negCount int64
				tx.Model(&Consultation{}).
					Where("application_id = ? AND response_type = 'negative'", c.ApplicationID).
					Count(&negCount)

				if negCount == 0 {
					// Advance status → gündemde
					var appStatus struct{ Status string }
					tx.Table("applications").Select("status").Where("id = ?", c.ApplicationID).First(&appStatus)

					if err := tx.Exec(
						"UPDATE applications SET status = ?, updated_at = ? WHERE id = ?",
						"gündemde", now, c.ApplicationID,
					).Error; err != nil {
						return fmt.Errorf("consultations: advance to gündemde: %w", err)
					}

					_ = writeLogTx(ctx, tx, "status.change", c.ApplicationID, "application", map[string]interface{}{
						"from":   appStatus.Status,
						"to":     "gündemde",
						"reason": "all consultations positive",
					})
				}
			}
		}

		return nil
	})
}

// ─── List ────────────────────────────────────────────────────────────────────

// ListForApplication returns all consultations for an application as summaries.
func (s *Service) ListForApplication(ctx context.Context, appID string) ([]*ConsultationSummary, error) {
	consultations, err := s.repo.FindByApplicationID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("consultations: list: %w", err)
	}

	summaries := make([]*ConsultationSummary, 0, len(consultations))
	for _, c := range consultations {
		status := "pending"
		if c.IsTokenExpired() && !c.IsTokenUsed() {
			status = "expired"
		} else if c.ResponseType != nil {
			status = string(*c.ResponseType)
		}

		var rt *string
		if c.ResponseType != nil {
			s := string(*c.ResponseType)
			rt = &s
		}

		summaries = append(summaries, &ConsultationSummary{
			ID:            c.ID,
			ApplicationID: c.ApplicationID,
			MemberUserID:  c.MemberUserID,
			MemberName:    c.MemberName,
			MemberEmail:   c.MemberEmail,
			ResponseType:  rt,
			Status:        status,
			CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return summaries, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (s *Service) writeLog(ctx context.Context, action, entityID, entityType string, meta map[string]interface{}) error {
	return writeLogTx(ctx, s.db, action, entityID, entityType, meta)
}

func writeLogTx(ctx context.Context, tx *gorm.DB, action, entityID, entityType string, meta map[string]interface{}) error {
	m, _ := json.Marshal(meta)
	entry := struct {
		ID         string         `gorm:"column:id"`
		Action     string         `gorm:"column:action"`
		ActorID    *string        `gorm:"column:actor_id"`
		ActorRole  string         `gorm:"column:actor_role"`
		EntityType string         `gorm:"column:entity_type"`
		EntityID   string         `gorm:"column:entity_id"`
		Metadata   datatypes.JSON `gorm:"column:metadata"`
		CreatedAt  time.Time      `gorm:"column:created_at"`
	}{
		ID:         uuid.New().String(),
		Action:     action,
		ActorRole:  "system",
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   datatypes.JSON(m),
		CreatedAt:  time.Now(),
	}
	return tx.WithContext(ctx).Table("logs").Create(&entry).Error
}
