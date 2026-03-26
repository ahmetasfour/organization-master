package notifications

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm"
)

// reminderRef is a lightweight projection used for the reminder cron query.
// It joins references with applications to get applicant name in one query.
type reminderRef struct {
	ID             string    `gorm:"column:id"`
	RefereeEmail   string    `gorm:"column:referee_email"`
	RefereeName    string    `gorm:"column:referee_name"`
	TokenExpiresAt time.Time `gorm:"column:token_expires_at"`
	// From applications join:
	ApplicantName string `gorm:"column:applicant_name"`
}

// ReminderJob holds state for the periodic reference-reminder cron.
type ReminderJob struct {
	db      *gorm.DB
	svc     *Service
	baseURL string
}

// NewReminderJob creates a new ReminderJob.
func NewReminderJob(db *gorm.DB, svc *Service, baseURL string) *ReminderJob {
	return &ReminderJob{db: db, svc: svc, baseURL: baseURL}
}

// Run executes one reminder cycle:
//  1. Find all unanswered references whose token expires within the next 24 hours.
//  2. Send a reminder email to each referee.
//  3. Log the action.
//
// This method is safe to call from a time.Ticker goroutine.
func (j *ReminderJob) Run(ctx context.Context) {
	now := time.Now()
	deadline := now.Add(24 * time.Hour)

	var refs []reminderRef
	err := j.db.WithContext(ctx).
		Table("references r").
		Select("r.id, r.referee_email, r.referee_name, r.token_expires_at, a.applicant_name").
		Joins("JOIN applications a ON a.id = r.application_id").
		Where("r.token_used_at IS NULL").
		Where("r.token_expires_at > ?", now).
		Where("r.token_expires_at <= ?", deadline).
		Scan(&refs).Error

	if err != nil {
		log.Printf("[CRON] reminder: query failed: %v", err)
		return
	}

	if len(refs) == 0 {
		return
	}

	log.Printf("[CRON] reminder: sending %d reminder(s)", len(refs))

	for _, ref := range refs {
		// The raw token is not stored in the DB — we cannot re-derive it.
		// Instead we build the respond URL using only the token hash path;
		// the frontend must accept hashed tokens on its respond page, OR
		// we generate a fresh token on resend (handled by ResendToken service method).
		// For cron reminders we use the existing token hash as the URL segment
		// because the referee already received the original raw token in their
		// first email. The cron just sends the original response URL again.
		//
		// Note: If the raw token is no longer available we send a reminder
		// directing the referee to contact the coordinator. For now we send
		// a generic reminder without a clickable link and log accordingly.
		if err := j.svc.SendReferenceReminder(
			ctx,
			ref.ID,
			ref.RefereeEmail,
			ref.RefereeName,
			"", // rawToken unavailable in cron — see note above
			ref.ApplicantName,
			ref.TokenExpiresAt,
		); err != nil {
			log.Printf("[CRON] reminder: failed to send to %s: %v", ref.RefereeEmail, err)
		}
	}
}

// Start launches the ReminderJob as a background goroutine that ticks every hour.
// It respects ctx cancellation for graceful shutdown.
func (j *ReminderJob) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		defer ticker.Stop()
		log.Println("[CRON] reminder: started (interval=1h)")
		for {
			select {
			case <-ticker.C:
				j.Run(ctx)
			case <-ctx.Done():
				log.Println("[CRON] reminder: stopped")
				return
			}
		}
	}()
}
