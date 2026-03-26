package consultations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConsultationResponseType represents the consulted member's response classification.
type ConsultationResponseType string

const (
	ConsultResponsePositive ConsultationResponseType = "positive"
	ConsultResponseNegative ConsultationResponseType = "negative"
)

// Consultation represents a single consultation record for a Profesyonel/Öğrenci application.
// Each record holds a single-use, 48-hour token sent to an existing system member.
type Consultation struct {
	ID            string `gorm:"type:char(36);primaryKey"       json:"id"`
	ApplicationID string `gorm:"type:char(36);not null;index"   json:"application_id"`

	// Member info (copied at creation time so emails/display don't need a JOIN)
	MemberUserID string `gorm:"type:char(36);not null;index"   json:"member_user_id"`
	MemberName   string `gorm:"type:varchar(255);not null"     json:"member_name"`
	MemberEmail  string `gorm:"type:varchar(255);not null"     json:"member_email"`

	// Token security — store SHA-256 hash only, never the raw token
	TokenHash      string     `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`
	TokenExpiresAt time.Time  `gorm:"not null;index"                        json:"token_expires_at"`
	TokenUsedAt    *time.Time `gorm:"type:timestamp null"                   json:"token_used_at,omitempty"`

	// Response (populated when member submits the form)
	ResponseType *ConsultationResponseType `gorm:"type:enum('positive','negative');index" json:"response_type,omitempty"`
	Reason       string                    `gorm:"type:text"                              json:"reason,omitempty"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM.
func (Consultation) TableName() string { return "consultations" }

// BeforeCreate GORM hook — assigns a UUID primary key if not set.
func (c *Consultation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// IsTokenExpired returns true if the token's 48-hour window has passed.
func (c *Consultation) IsTokenExpired() bool { return time.Now().After(c.TokenExpiresAt) }

// IsTokenUsed returns true if the token has already been consumed.
func (c *Consultation) IsTokenUsed() bool { return c.TokenUsedAt != nil }

// HasResponse returns true if the member has already submitted a response.
func (c *Consultation) HasResponse() bool { return c.ResponseType != nil }
