package reputation

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReputationResponseType represents the reputation contact response classification
type ReputationResponseType string

const (
	ReputationClean    ReputationResponseType = "clean"
	ReputationNegative ReputationResponseType = "negative"
)

// ReputationContact represents a reputation screening contact for Asil/Akademik applications
type ReputationContact struct {
	ID                  string `gorm:"type:char(36);primaryKey" json:"id"`
	ApplicationID       string `gorm:"type:char(36);not null;index" json:"application_id"`
	ContactName         string `gorm:"type:varchar(255);not null" json:"contact_name"`
	ContactEmail        string `gorm:"type:varchar(255)" json:"contact_email,omitempty"`
	ContactPhone        string `gorm:"type:varchar(50)" json:"contact_phone,omitempty"`
	ContactOrganization string `gorm:"type:varchar(255)" json:"contact_organization,omitempty"`

	ResponseType ReputationResponseType `gorm:"type:enum('clean','negative');not null;index" json:"response_type"`
	Notes        string                 `gorm:"type:text" json:"notes,omitempty"`

	ContactedByUserID *string `gorm:"type:char(36);index" json:"contacted_by_user_id,omitempty"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (ReputationContact) TableName() string {
	return "reputation_contacts"
}

// BeforeCreate GORM hook to generate UUID before creating a reputation contact
func (rc *ReputationContact) BeforeCreate(tx *gorm.DB) error {
	if rc.ID == "" {
		rc.ID = uuid.New().String()
	}
	return nil
}

// IsClean checks if the reputation contact response is clean
func (rc *ReputationContact) IsClean() bool {
	return rc.ResponseType == ReputationClean
}

// IsNegative checks if the reputation contact response is negative
func (rc *ReputationContact) IsNegative() bool {
	return rc.ResponseType == ReputationNegative
}
