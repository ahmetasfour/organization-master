package references

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

// Service contains the business logic for the reference system.
type Service struct {
	repo      *Repository
	authRepo  *auth.Repository
	logRepo   *logs.Repository
	notifySvc *notifications.Service
	db        *gorm.DB
}

// NewService creates a new reference service.
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

// ─── CreateForApplication ──────────────────────────────────────────────────────

// CreateForApplication creates one Reference record per referee, sends tokenized
// emails, and advances the application status to referans_bekleniyor.
//
// This method satisfies the applications.ReferenceCreator interface.
func (s *Service) CreateForApplication(
	ctx context.Context,
	appID, applicantName, applicantEmail, membershipType string,
	refereeUserIDs []string,
) error {
	app := AppContext{
		ID:             appID,
		ApplicantName:  applicantName,
		ApplicantEmail: applicantEmail,
		MembershipType: membershipType,
	}
	return s.createForApp(ctx, app, refereeUserIDs)
}

// createForApp is the internal implementation shared by CreateForApplication.
func (s *Service) createForApp(ctx context.Context, app AppContext, refereeUserIDs []string) error {
	refs := make([]Reference, 0, len(refereeUserIDs))

	for _, userID := range refereeUserIDs {
		// Look up user — must exist and be active
		user, err := s.authRepo.FindByID(ctx, userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("references: referee user %s not found", userID)
			}
			return fmt.Errorf("references: lookup referee: %w", err)
		}
		if !user.IsActive {
			return fmt.Errorf("references: referee user %s is inactive", userID)
		}

		tok := shared.GenerateToken()

		refs = append(refs, Reference{
			ApplicationID:  app.ID,
			RefereeName:    user.FullName,
			RefereeEmail:   user.Email,
			TokenHash:      tok.HashedToken,
			TokenExpiresAt: tok.ExpiresAt,
			IsReplacement:  false,
			Round:          1,
		})

		// Send email (use raw token in URL — never the hash)
		if err := s.notifySvc.SendReferenceRequest(
			ctx,
			"", // refID not yet set — will be logged with appID
			user.Email,
			user.FullName,
			tok.RawToken,
			app.ApplicantName,
			string(app.MembershipType),
			tok.ExpiresAt,
		); err != nil {
			// Non-fatal: log and continue so other refs still get emails
			_ = s.writeLog(ctx, "ref.email_failed", app.ID, "application", map[string]interface{}{
				"referee_email": user.Email,
				"error":         err.Error(),
			})
		}
	}

	// Persist all reference records
	if err := s.repo.CreateBatch(ctx, refs); err != nil {
		return fmt.Errorf("references: create batch: %w", err)
	}

	// Log bulk send
	_ = s.writeLog(ctx, "ref.sent", app.ID, "application", map[string]interface{}{
		"count": len(refs),
	})

	// Advance application status → referans_bekleniyor
	if err := s.db.WithContext(ctx).
		Exec("UPDATE applications SET status = ?, updated_at = ? WHERE id = ?",
			"referans_bekleniyor", time.Now(), app.ID).Error; err != nil {
		return fmt.Errorf("references: advance status: %w", err)
	}

	return nil
}

// ─── GetFormData ───────────────────────────────────────────────────────────────

// GetFormData retrieves the public form data for a given raw token.
// Returns ErrTokenExpired or ErrTokenUsed without consuming the token.
func (s *Service) GetFormData(ctx context.Context, rawToken string) (*ReferenceFormData, error) {
	hash := shared.HashToken(rawToken)

	ref, err := s.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("references: get form data: %w", err)
	}

	if ref.IsTokenExpired() {
		return nil, shared.ErrTokenExpired
	}
	if ref.IsTokenUsed() {
		return nil, shared.ErrTokenUsed
	}

	// Load applicant name from application row (lightweight query)
	type appRow struct {
		ApplicantName  string `gorm:"column:applicant_name"`
		MembershipType string `gorm:"column:membership_type"`
	}
	var row appRow
	if err := s.db.WithContext(ctx).
		Table("applications").
		Select("applicant_name", "membership_type").
		Where("id = ?", ref.ApplicationID).
		First(&row).Error; err != nil {
		return nil, fmt.Errorf("references: load application: %w", err)
	}

	return &ReferenceFormData{
		ApplicantName:  row.ApplicantName,
		MembershipType: row.MembershipType,
		RefereeName:    ref.RefereeName,
		ExpiresAt:      notifications.FormatTime(ref.TokenExpiresAt),
	}, nil
}

// ─── SubmitResponse ────────────────────────────────────────────────────────────

// SubmitResponse validates and consumes a reference token, saves the referee's
// response, and applies the appropriate business logic:
//   - negative → RedGuard.Terminate
//   - unknown  → notify applicant, create replacement reference slot
//   - positive → check if all references are now complete → advance status
func (s *Service) SubmitResponse(
	ctx context.Context,
	rawToken string,
	req *ReferenceResponseRequest,
	ipAddress string,
) error {
	// Validate reason for negative
	if req.ResponseType == string(ResponseNegative) && len(req.Reason) < 30 {
		return fmt.Errorf("reason: minimum 30 characters required for negative response")
	}

	hash := shared.HashToken(rawToken)
	now := time.Now()

	// Atomic token validation + consume inside a transaction
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Validate and consume token
		if err := shared.ValidateAndConsumeToken(ctx, tx, "references", hash, now); err != nil {
			return err
		}

		// 2. Reload reference within transaction
		var ref Reference
		if err := tx.Where("token_hash = ?", hash).First(&ref).Error; err != nil {
			return fmt.Errorf("references: reload ref: %w", err)
		}

		// 3. Save response
		resp := &ReferenceResponse{
			ID:                 uuid.New().String(),
			ReferenceID:        ref.ID,
			ResponseType:       ResponseType(req.ResponseType),
			RecommendationText: req.Reason,
			CreatedAt:          now,
		}
		if err := tx.Create(resp).Error; err != nil {
			return fmt.Errorf("references: save response: %w", err)
		}

		// 4. Log the response
		_ = writeLogTx(ctx, tx, "ref.responded", ref.ApplicationID, "application", map[string]interface{}{
			"response_type": req.ResponseType,
			"ip":            ipAddress,
			"ref_id":        ref.ID,
		})

		// 5. Apply business logic per response type
		switch ResponseType(req.ResponseType) {
		case ResponseNegative:
			redGuard := shared.NewRedGuard(tx)
			if err := redGuard.Terminate(ctx, ref.ApplicationID, req.Reason, "system", "referee"); err != nil {
				if errors.Is(err, shared.ErrApplicationTerminated) {
					return nil // already terminated — idempotent
				}
				return fmt.Errorf("references: terminate: %w", err)
			}

			// Send rejection email to applicant (reason NOT included)
			type appInfo struct {
				ApplicantName  string `gorm:"column:applicant_name"`
				ApplicantEmail string `gorm:"column:applicant_email"`
			}
			var info appInfo
			if err := tx.Table("applications").
				Select("applicant_name", "applicant_email").
				Where("id = ?", ref.ApplicationID).
				First(&info).Error; err == nil {
				_ = s.notifySvc.SendRejected(ctx, ref.ApplicationID, info.ApplicantEmail, info.ApplicantName)
			}

		case ResponseUnknown:
			// Notify applicant to add a replacement reference
			type appInfo2 struct {
				ApplicantName  string `gorm:"column:applicant_name"`
				ApplicantEmail string `gorm:"column:applicant_email"`
			}
			var info appInfo2
			if err := tx.Table("applications").
				Select("applicant_name", "applicant_email").
				Where("id = ?", ref.ApplicationID).
				First(&info).Error; err != nil {
				return fmt.Errorf("references: load app for unknown: %w", err)
			}

			// Determine next round
			var maxRound struct{ Round int }
			tx.Table("references").
				Select("MAX(round) AS round").
				Where("application_id = ?", ref.ApplicationID).
				Scan(&maxRound)
			nextRound := maxRound.Round + 1

			// Create replacement reference placeholder (no token yet — coordinator must resend)
			replacement := &Reference{
				ApplicationID:  ref.ApplicationID,
				RefereeName:    "", // to be filled when applicant selects new referee
				RefereeEmail:   "",
				TokenHash:      uuid.New().String(),           // placeholder — will be overwritten on resend
				TokenExpiresAt: now.Add(1 * time.Millisecond), // expired immediately so it can't be used
				IsReplacement:  true,
				Round:          nextRound,
			}
			if err := tx.Create(replacement).Error; err != nil {
				return fmt.Errorf("references: create replacement: %w", err)
			}

			_ = writeLogTx(ctx, tx, "ref.replacement_requested", ref.ApplicationID, "application", map[string]interface{}{
				"unknown_referee": ref.RefereeName,
				"round":           nextRound,
			})

			// Notify applicant
			_ = s.notifySvc.SendNewRefNeeded(ctx, ref.ApplicationID, info.ApplicantEmail, info.ApplicantName, ref.RefereeName)

		case ResponsePositive:
			// Check if ALL references are done and we have >= 3 positives
			var pending int64
			tx.Model(&Reference{}).
				Where("application_id = ? AND token_used_at IS NULL", ref.ApplicationID).
				Count(&pending)

			if pending == 0 {
				var positiveCount int64
				tx.Model(&ReferenceResponse{}).
					Joins("JOIN `references` ON `references`.id = reference_responses.reference_id").
					Where("`references`.application_id = ? AND reference_responses.response_type = 'positive'", ref.ApplicationID).
					Count(&positiveCount)

				if positiveCount >= 3 {
					// Load current status to validate transition
					var appStatus struct {
						Status string `gorm:"column:status"`
					}
					tx.Table("applications").Select("status").Where("id = ?", ref.ApplicationID).First(&appStatus)

					if err := tx.Exec(
						"UPDATE applications SET status = ?, updated_at = ? WHERE id = ?",
						"referans_tamamlandı", now, ref.ApplicationID,
					).Error; err != nil {
						return fmt.Errorf("references: advance to tamamlandı: %w", err)
					}

					_ = writeLogTx(ctx, tx, "status.change", ref.ApplicationID, "application", map[string]interface{}{
						"from":   appStatus.Status,
						"to":     "referans_tamamlandı",
						"reason": "all references positive",
					})
				}
			}
		}

		return nil
	})
}

// ─── ResendToken ───────────────────────────────────────────────────────────────

// ResendToken regenerates and resends a reference request email for a given reference ID.
func (s *Service) ResendToken(
	ctx context.Context,
	refID string,
	refereeName, refereeEmail string,
	applicantName, membershipType string,
) error {
	// Regenerate token
	tok := shared.GenerateToken()

	if err := s.repo.UpdateToken(ctx, refID, tok.HashedToken, tok.ExpiresAt); err != nil {
		return fmt.Errorf("references: update token: %w", err)
	}

	// Resend email with new raw token
	if err := s.notifySvc.SendReferenceRequest(
		ctx,
		refID,
		refereeEmail,
		refereeName,
		tok.RawToken,
		applicantName,
		membershipType,
		tok.ExpiresAt,
	); err != nil {
		return fmt.Errorf("references: resend email: %w", err)
	}

	_ = s.writeLog(ctx, "ref.resent", refID, "reference", map[string]interface{}{
		"referee_email": refereeEmail,
	})

	return nil
}

// ─── helpers ───────────────────────────────────────────────────────────────────

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
