package reputation

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReputationResponseType represents the reputation contact response classification.
type ReputationResponseType string

const (
	ResponseClean    ReputationResponseType = "clean"
	ResponseNegative ReputationResponseType = "negative"
)

// ReputationContact represents a tokenized reputation-screening contact sent for an
// Asil or Akademik membership application. Each contact receives a unique one-time
// token URL and submits their response through a public page.
type ReputationContact struct {
	ID            string `gorm:"type:char(36);primaryKey" json:"id"`
	ApplicationID string `gorm:"type:char(36);not null;index" json:"application_id"`
	ContactName   string `gorm:"type:varchar(255);not null" json:"contact_name"`
	ContactEmail  string `gorm:"type:varchar(255);not null;index" json:"contact_email"`

	// Token fields — store SHA-256 hash only, raw token goes in email URL.
	TokenHash      string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`
	TokenExpiresAt time.Time  `gorm:"not null;index" json:"token_expires_at"`
	TokenUsedAt    *time.Time `gorm:"type:timestamp null" json:"token_used_at,omitempty"`

	// Response fields — set when the contact submits their response.
	ResponseType *ReputationResponseType `gorm:"type:enum('clean','negative');index" json:"response_type,omitempty"`
	Reason       string                  `gorm:"type:text" json:"reason,omitempty"`
	RespondedAt  *time.Time              `gorm:"type:timestamp null" json:"responded_at,omitempty"`
	RespondedIP  string                  `gorm:"type:varchar(45)" json:"-"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (ReputationContact) TableName() string { return "reputation_contacts" }

// BeforeCreate generates a UUID primary key if one is not already set.
func (rc *ReputationContact) BeforeCreate(tx *gorm.DB) error {
	if rc.ID == "" {
		rc.ID = uuid.New().String()
	}
	return nil
}

// IsTokenExpired returns true when the token's validity window has passed.
func (rc *ReputationContact) IsTokenExpired() bool {
	return time.Now().After(rc.TokenExpiresAt)
}

// IsTokenUsed returns true when the contact has already submitted a response.
func (rc *ReputationContact) IsTokenUsed() bool {
	return rc.TokenUsedAt != nil
}

// IsResponded returns true when the contact submitted any response.
func (rc *ReputationContact) IsResponded() bool {
	return rc.ResponseType != nil
}

// IsClean returns true when the contact reported no negative information.
func (rc *ReputationContact) IsClean() bool {
	return rc.ResponseType != nil && *rc.ResponseType == ResponseClean
}

// IsNegative returns true when the contact reported negative information.
func (rc *ReputationContact) IsNegative() bool {
	return rc.ResponseType != nil && *rc.ResponseType == ResponseNegative
}
