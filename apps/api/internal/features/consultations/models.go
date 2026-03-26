package consultations

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Consultation represents a consultation process for Profesyonel/Öğrenci applications
type Consultation struct {
	ID               string `gorm:"type:char(36);primaryKey" json:"id"`
	ApplicationID    string `gorm:"type:char(36);not null;index" json:"application_id"`
	AssignedToUserID string `gorm:"type:char(36);not null;index" json:"assigned_to_user_id"`

	Notes          string `gorm:"type:text" json:"notes,omitempty"`
	Recommendation string `gorm:"type:text" json:"recommendation,omitempty"`
	IsApproved     *bool  `gorm:"type:boolean" json:"is_approved,omitempty"`

	CompletedAt *time.Time `gorm:"type:timestamp null;index" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Consultation) TableName() string {
	return "consultations"
}

// BeforeCreate GORM hook to generate UUID before creating a consultation
func (c *Consultation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// IsCompleted checks if the consultation has been completed
func (c *Consultation) IsCompleted() bool {
	return c.CompletedAt != nil
}

// IsApprovedConsultation checks if the consultation resulted in approval
func (c *Consultation) IsApprovedConsultation() bool {
	return c.IsApproved != nil && *c.IsApproved
}

// IsRejectedConsultation checks if the consultation resulted in rejection
func (c *Consultation) IsRejectedConsultation() bool {
	return c.IsApproved != nil && !*c.IsApproved
}
