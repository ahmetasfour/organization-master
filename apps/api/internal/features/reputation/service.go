package reputation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/features/notifications"
	"membership-system/api/internal/shared"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// repAppRow is a minimal projection of the applications table used to avoid
// importing the applications package (prevents circular dependency).
type repAppRow struct {
	ID             string `gorm:"column:id"`
	MembershipType string `gorm:"column:membership_type"`
	Status         string `gorm:"column:status"`
	ApplicantName  string `gorm:"column:applicant_name"`
	ApplicantEmail string `gorm:"column:applicant_email"`
	LinkedInURL    string `gorm:"column:linkedin_url"`
}

func (repAppRow) TableName() string { return "applications" }

// reputationTypes are the only membership types that go through reputation screening.
var reputationTypes = map[string]bool{
	"asil":     true,
	"akademik": true,
}

// Service contains business logic for the reputation screening system.
type Service struct {
	repo      *Repository
	authRepo  *auth.Repository
	logRepo   *logs.Repository
	notifySvc *notifications.Service
	db        *gorm.DB
}

// NewService creates a new reputation service.
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

// ─── AddContacts ──────────────────────────────────────────────────────────────

// AddContacts validates, creates tokenized reputation contact records, sends
// query emails, and advances the application status → itibar_taramasında.
//
// CRITICAL: Exactly 10 contacts must be provided.
func (s *Service) AddContacts(
	ctx context.Context,
	appID string,
	req *AddContactsRequest,
	actorID, actorRole string,
) error {
	// Validate exactly 10 contacts (enforced at both validator and service layers)
	if len(req.Contacts) != 10 {
		return fmt.Errorf("itibar: tam olarak 10 kişi gereklidir, %d kişi verildi", len(req.Contacts))
	}

	// Load application
	var app repAppRow
	if err := s.db.WithContext(ctx).First(&app, "id = ?", appID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrNotFound
		}
		return fmt.Errorf("itibar: başvuru yüklenemedi: %w", err)
	}

	// Validate membership type
	if !reputationTypes[app.MembershipType] {
		return fmt.Errorf("itibar: itibar taraması sadece asil/akademik başvurular için geçerlidir, mevcut: %s", app.MembershipType)
	}

	// Validate application status must be ön_onaylandı
	if app.Status != "ön_onaylandı" {
		return fmt.Errorf("itibar: itibar kişileri eklemek için başvuru ön_onaylandı durumunda olmalıdır, mevcut: %s", app.Status)
	}

	// RedGuard: must not be terminated
	redGuard := shared.NewRedGuard(s.db)
	if err := redGuard.AssertNotTerminated(ctx, appID); err != nil {
		return err
	}

	// Build contact records
	contacts := make([]ReputationContact, 0, len(req.Contacts))
	type pendingEmail struct {
		contact  ReputationContact
		rawToken string
	}
	pending := make([]pendingEmail, 0, len(req.Contacts))

	for _, input := range req.Contacts {
		tok := shared.GenerateToken()
		c := ReputationContact{
			ApplicationID:  appID,
			ContactName:    input.Name,
			ContactEmail:   input.Email,
			TokenHash:      tok.HashedToken,
			TokenExpiresAt: tok.ExpiresAt,
		}
		contacts = append(contacts, c)
		pending = append(pending, pendingEmail{contact: c, rawToken: tok.RawToken})
	}

	// Persist all contacts
	if err := s.repo.CreateBatch(ctx, contacts); err != nil {
		return fmt.Errorf("itibar: kişiler kaydedilemedi: %w", err)
	}

	// Send emails (non-fatal per contact)
	for _, p := range pending {
		if err := s.notifySvc.SendReputationQuery(
			ctx,
			p.contact.ID,
			p.contact.ContactEmail,
			p.contact.ContactName,
			p.rawToken,
			app.ApplicantName,
			app.LinkedInURL,
			p.contact.TokenExpiresAt,
		); err != nil {
			_ = s.writeLog(ctx, "rep.email_failed", appID, "application", map[string]interface{}{
				"contact_email": p.contact.ContactEmail,
				"error":         err.Error(),
			})
		} else {
			_ = s.writeLog(ctx, "rep.contact_added", appID, "application", map[string]interface{}{
				"contact_name":  p.contact.ContactName,
				"contact_email": p.contact.ContactEmail,
			})
		}
	}

	// Advance status → itibar_taramasında
	if err := s.db.WithContext(ctx).
		Exec("UPDATE applications SET status = ?, updated_at = ? WHERE id = ?",
			"itibar_taramasında", time.Now(), appID).Error; err != nil {
		return fmt.Errorf("itibar: durum güncellenemedi: %w", err)
	}

	_ = s.writeLog(ctx, "status.change", appID, "application", map[string]interface{}{
		"from":   "ön_onaylandı",
		"to":     "itibar_taramasında",
		"actor":  actorID,
		"reason": "reputation contacts added",
	})

	return nil
}

// ─── GetFormData ──────────────────────────────────────────────────────────────

// GetFormData retrieves public context for a raw reputation token.
// Does NOT consume the token — only verifies expiry/used status.
func (s *Service) GetFormData(ctx context.Context, rawToken string) (*ReputationFormData, error) {
	hash := shared.HashToken(rawToken)

	c, err := s.repo.FindByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("itibar: token bulunamadı: %w", err)
	}

	if c.IsTokenExpired() {
		return nil, shared.ErrTokenExpired
	}
	if c.IsTokenUsed() {
		return nil, shared.ErrTokenUsed
	}

	// Load application context
	var app repAppRow
	if err := s.db.WithContext(ctx).First(&app, "id = ?", c.ApplicationID).Error; err != nil {
		return nil, fmt.Errorf("itibar: başvuru yüklenemedi: %w", err)
	}

	return &ReputationFormData{
		ContactName:       c.ContactName,
		ApplicantName:     app.ApplicantName,
		ApplicantLinkedIn: app.LinkedInURL,
		ExpiresAt:         notifications.FormatTime(c.TokenExpiresAt),
	}, nil
}

// ─── SubmitResponse ───────────────────────────────────────────────────────────

// SubmitResponse validates and consumes a reputation token, saves the contact's
// response, and applies the appropriate business logic:
//   - negative → notify YK members (do NOT auto-terminate; YK must vote)
//   - clean    → if all 10 responded clean → advance to itibar_temiz
func (s *Service) SubmitResponse(
	ctx context.Context,
	rawToken string,
	req *ContactResponseRequest,
	ipAddress string,
) error {
	if req.ResponseType == string(ResponseNegative) && len(req.Reason) < 30 {
		return fmt.Errorf("gerekçe: olumsuz yanıtlar için minimum 30 karakter gereklidir")
	}

	hash := shared.HashToken(rawToken)
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Validate and atomically consume token
		if err := shared.ValidateAndConsumeToken(ctx, tx, "reputation_contacts", hash, now); err != nil {
			return err
		}

		// 2. Reload contact within transaction
		var c ReputationContact
		if err := tx.Where("token_hash = ?", hash).First(&c).Error; err != nil {
			return fmt.Errorf("itibar: kişi yüklenemedi: %w", err)
		}

		// 3. Save response
		respType := ReputationResponseType(req.ResponseType)
		if err := tx.Model(&ReputationContact{}).
			Where("id = ?", c.ID).
			Updates(map[string]interface{}{
				"response_type": respType,
				"reason":        req.Reason,
				"responded_at":  now,
				"responded_ip":  ipAddress,
			}).Error; err != nil {
			return fmt.Errorf("itibar: yanıt kaydedilemedi: %w", err)
		}

		// 4. Audit log
		_ = repWriteLogTx(ctx, tx, "rep.responded", c.ApplicationID, "application", map[string]interface{}{
			"response_type": req.ResponseType,
			"ip":            ipAddress,
			"contact_id":    c.ID,
		})

		// 5. Business logic
		switch respType {
		case ResponseNegative:
			// Per spec: do NOT auto-terminate on negative reputation response.
			// Instead: notify all YK members so they can review and vote.
			var ykUsers []struct {
				ID       string `gorm:"column:id"`
				Email    string `gorm:"column:email"`
				FullName string `gorm:"column:full_name"`
			}
			if err := tx.Table("users").
				Select("id, email, full_name").
				Where("role = 'yk' AND is_active = true").
				Find(&ykUsers).Error; err == nil && len(ykUsers) > 0 {

				ykList := make([]struct {
					ID    string
					Email string
					Name  string
				}, len(ykUsers))
				for i, u := range ykUsers {
					ykList[i] = struct {
						ID    string
						Email string
						Name  string
					}{ID: u.ID, Email: u.Email, Name: u.FullName}
				}

				// Load application for notification context
				var app repAppRow
				if loadErr := tx.Table("applications").
					Where("id = ?", c.ApplicationID).
					First(&app).Error; loadErr == nil {
					_ = s.notifySvc.SendHonoraryProposal(
						ctx,
						c.ApplicationID,
						"Sistem",
						app.ApplicantName,
						app.LinkedInURL,
						fmt.Sprintf("İtibar taramasında olumsuz geri dönüş alındı. İnceleme gerekiyor.\n\nİletişim kişisi: %s", c.ContactName),
						ykList,
					)
				}
			}

		case ResponseClean:
			// Check if all 10 contacts have responded and none are negative
			var total, responded, negCount int64
			tx.Model(&ReputationContact{}).
				Where("application_id = ?", c.ApplicationID).
				Count(&total)
			tx.Model(&ReputationContact{}).
				Where("application_id = ? AND response_type IS NOT NULL", c.ApplicationID).
				Count(&responded)
			tx.Model(&ReputationContact{}).
				Where("application_id = ? AND response_type = 'negative'", c.ApplicationID).
				Count(&negCount)

			if responded == total && negCount == 0 {
				// All responded clean → advance to itibar_temiz
				var appStatus struct{ Status string }
				tx.Table("applications").Select("status").Where("id = ?", c.ApplicationID).First(&appStatus)

				if err := tx.Exec(
					"UPDATE applications SET status = ?, updated_at = ? WHERE id = ?",
					"itibar_temiz", now, c.ApplicationID,
				).Error; err != nil {
					return fmt.Errorf("itibar: durum güncellenemedi (itibar_temiz): %w", err)
				}

				_ = repWriteLogTx(ctx, tx, "status.change", c.ApplicationID, "application", map[string]interface{}{
					"from":   appStatus.Status,
					"to":     "itibar_temiz",
					"reason": "all 10 reputation contacts responded clean",
				})
			}
		}

		return nil
	})
}

// ─── GetStatus ────────────────────────────────────────────────────────────────

// GetStatus returns the aggregated reputation screening status for an application.
// Contact email addresses are always masked: j***@example.com
func (s *Service) GetStatus(ctx context.Context, appID string) (*ReputationStatusResponse, error) {
	contacts, err := s.repo.FindByApplicationID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("itibar: durum sorgulanamadı: %w", err)
	}

	resp := &ReputationStatusResponse{
		ApplicationID: appID,
		TotalContacts: len(contacts),
		Contacts:      make([]ContactStatus, 0, len(contacts)),
	}

	for _, c := range contacts {
		status := "pending"
		if c.ResponseType != nil {
			switch *c.ResponseType {
			case ResponseClean:
				status = "clean"
				resp.Clean++
				resp.Responded++
			case ResponseNegative:
				status = "flagged"
				resp.Flagged++
				resp.Responded++
			}
		}

		resp.Contacts = append(resp.Contacts, ContactStatus{
			ID:          c.ID,
			ContactName: c.ContactName,
			Email:       maskEmail(c.ContactEmail),
			Status:      status,
			RespondedAt: c.RespondedAt,
		})
	}

	return resp, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// maskEmail returns a privacy-safe masked version of an email: j***@example.com
func maskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || len(parts[0]) == 0 {
		return "***"
	}
	return string(parts[0][0]) + "***@" + parts[1]
}

func (s *Service) writeLog(ctx context.Context, action, entityID, entityType string, meta map[string]interface{}) error {
	return repWriteLogTx(ctx, s.db, action, entityID, entityType, meta)
}

func repWriteLogTx(ctx context.Context, tx *gorm.DB, action, entityID, entityType string, meta map[string]interface{}) error {
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
