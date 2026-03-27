package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RedGuard enforces the immutable-once-terminated rule on applications.
// All termination must go through this guard to ensure consistency.
type RedGuard struct {
	db *gorm.DB
}

// NewRedGuard creates a new RedGuard with the given database connection.
func NewRedGuard(db *gorm.DB) *RedGuard {
	return &RedGuard{db: db}
}

// applicationRow is a minimal struct used for loading application state.
type applicationRow struct {
	ID              string  `gorm:"column:id"`
	Status          string  `gorm:"column:status"`
	RejectionReason *string `gorm:"column:rejection_reason"`
	ApplicantEmail  string  `gorm:"column:applicant_email"`
	ApplicantName   string  `gorm:"column:applicant_name"`
}

func (applicationRow) TableName() string { return "applications" }

// logRow is used for inserting audit log entries.
type logRow struct {
	ID         string         `gorm:"column:id"`
	Action     string         `gorm:"column:action"`
	ActorID    *string        `gorm:"column:actor_id"`
	ActorRole  string         `gorm:"column:actor_role"`
	EntityType string         `gorm:"column:entity_type"`
	EntityID   string         `gorm:"column:entity_id"`
	Metadata   datatypes.JSON `gorm:"column:metadata"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
}

func (logRow) TableName() string { return "logs" }

// terminalStatuses mirrors the applications package to avoid a circular import.
var terminalStatusSet = map[string]bool{
	"referans_red": true,
	"yk_red":       true,
	"itibar_red":   true,
	"danışma_red":  true,
	"yik_red":      true,
	"reddedildi":   true,
	"kabul":        true,
}

// Terminate permanently terminates an application within a single DB transaction.
// It sets status = reddedildi, records rejection_reason (write-once), and writes an audit log.
func (rg *RedGuard) Terminate(ctx context.Context, applicationID, reason, actorID, actorRole string) error {
	return rg.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Load application — lock the row for update
		var app applicationRow
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&app, "id = ?", applicationID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("redguard: başvuru yüklenemedi: %w", err)
		}

		// 2. Check not already terminated
		if terminalStatusSet[app.Status] {
			return ErrApplicationTerminated
		}

		// 3. Build update map — rejection_reason is write-once (never overwrite)
		updates := map[string]interface{}{
			"status":           "reddedildi",
			"rejected_by_role": actorRole,
			"updated_at":       time.Now(),
		}
		if app.RejectionReason == nil || *app.RejectionReason == "" {
			updates["rejection_reason"] = reason
		}

		if err := tx.Table("applications").
			Where("id = ?", applicationID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("redguard: başvuru güncellenemedi: %w", err)
		}

		// 4. Write audit log
		metadata, _ := json.Marshal(map[string]interface{}{
			"reason":         reason,
			"actorRole":      actorRole,
			"previousStatus": app.Status,
			"applicantEmail": app.ApplicantEmail,
		})

		var actorIDPtr *string
		if actorID != "" {
			actorIDPtr = &actorID
		}

		entry := logRow{
			ID:         uuid.New().String(),
			Action:     "application.terminated",
			ActorID:    actorIDPtr,
			ActorRole:  actorRole,
			EntityType: "application",
			EntityID:   applicationID,
			Metadata:   datatypes.JSON(metadata),
			CreatedAt:  time.Now(),
		}
		if err := tx.Create(&entry).Error; err != nil {
			return fmt.Errorf("redguard: log yazılamadı: %w", err)
		}

		return nil
	})
}

// AssertNotTerminated returns ErrApplicationTerminated if the application is
// in any terminal state. Returns ErrNotFound if the application does not exist.
func (rg *RedGuard) AssertNotTerminated(ctx context.Context, applicationID string) error {
	terminated, err := rg.IsTerminated(ctx, applicationID)
	if err != nil {
		return err
	}
	if terminated {
		return ErrApplicationTerminated
	}
	return nil
}

// IsTerminated returns true if the application is in a terminal state.
func (rg *RedGuard) IsTerminated(ctx context.Context, applicationID string) (bool, error) {
	var app applicationRow
	if err := rg.db.WithContext(ctx).Select("id", "status").
		First(&app, "id = ?", applicationID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("redguard: sonlandırma kontrolü başarısız: %w", err)
	}
	return terminalStatusSet[app.Status], nil
}
