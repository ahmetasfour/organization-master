package applications

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MembershipType represents all membership categories
type MembershipType string

const (
	MembershipAsil        MembershipType = "asil"
	MembershipAkademik    MembershipType = "akademik"
	MembershipProfesyonel MembershipType = "profesyonel"
	MembershipOgrenci     MembershipType = "öğrenci"
	MembershipOnursal     MembershipType = "onursal"
)

// ApplicationStatus represents all possible states in the application state machine
type ApplicationStatus string

const (
	// Common statuses
	StatusBasvuruAlindi ApplicationStatus = "başvuru_alındı"
	StatusGundemde      ApplicationStatus = "gündemde"
	StatusKabul         ApplicationStatus = "kabul"
	StatusReddedildi    ApplicationStatus = "reddedildi"

	// Asil & Akademik flow
	StatusReferansBekleniyor ApplicationStatus = "referans_bekleniyor"
	StatusReferansTamamlandi ApplicationStatus = "referans_tamamlandı"
	StatusReferansRed        ApplicationStatus = "referans_red"
	StatusYKOnIncelemede     ApplicationStatus = "yk_ön_incelemede"
	StatusOnOnaylandi        ApplicationStatus = "ön_onaylandı"
	StatusYKRed              ApplicationStatus = "yk_red"
	StatusItibarTaramasinda  ApplicationStatus = "itibar_taramasında"
	StatusItibarTemiz        ApplicationStatus = "itibar_temiz"
	StatusItibarRed          ApplicationStatus = "itibar_red"

	// Profesyonel & Öğrenci flow
	StatusDanismaSurecinde ApplicationStatus = "danışma_sürecinde"
	StatusDanismaRed       ApplicationStatus = "danışma_red"

	// Onursal flow
	StatusOneriAlindi        ApplicationStatus = "öneri_alındı"
	StatusYIKDegerlendirmede ApplicationStatus = "yik_değerlendirmede"
	StatusYIKRed             ApplicationStatus = "yik_red"
)

// Application represents a membership application with full state tracking
type Application struct {
	ID             string            `gorm:"type:char(36);primaryKey" json:"id"`
	ApplicantName  string            `gorm:"type:varchar(255);not null" json:"applicant_name"`
	ApplicantEmail string            `gorm:"type:varchar(255);not null;index" json:"applicant_email"`
	ApplicantPhone string            `gorm:"type:varchar(50)" json:"applicant_phone,omitempty"`
	ApplicantBio   string            `gorm:"type:text" json:"applicant_bio,omitempty"`
	LinkedInURL    string            `gorm:"type:varchar(500)" json:"linkedin_url,omitempty"`
	PhotoURL       string            `gorm:"type:varchar(500)" json:"photo_url,omitempty"`
	MembershipType MembershipType    `gorm:"type:enum('asil','akademik','profesyonel','öğrenci','onursal');not null;index" json:"membership_type"`
	Status         ApplicationStatus `gorm:"type:enum('başvuru_alındı','referans_bekleniyor','referans_tamamlandı','referans_red','yk_ön_incelemede','ön_onaylandı','yk_red','itibar_taramasında','itibar_temiz','itibar_red','danışma_sürecinde','danışma_red','öneri_alındı','yik_değerlendirmede','yik_red','gündemde','kabul','reddedildi');not null;default:'başvuru_alındı';index" json:"status"`

	// Onursal-specific fields
	ProposedByUserID *string `gorm:"type:char(36);index" json:"proposed_by_user_id,omitempty"`
	ProposalReason   string  `gorm:"type:text" json:"proposal_reason,omitempty"`

	// Rejection tracking (WRITE-ONCE field - immutable once set)
	RejectionReason *string `gorm:"type:text" json:"rejection_reason,omitempty"`
	RejectedByRole  string  `gorm:"type:varchar(50)" json:"rejected_by_role,omitempty"`

	// Web publishing
	WebPublishConsent bool `gorm:"default:false;not null" json:"web_publish_consent"`
	IsPublished       bool `gorm:"default:false;not null" json:"is_published"`

	// Re-application tracking
	PreviousAppID *string `gorm:"type:char(36)" json:"previous_app_id,omitempty"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Application) TableName() string {
	return "applications"
}

// BeforeCreate GORM hook to generate UUID before creating an application
func (a *Application) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}

// IsTerminated checks if the application is in a terminal state
func (a *Application) IsTerminated() bool {
	switch a.Status {
	case StatusReferansRed, StatusYKRed, StatusItibarRed, StatusDanismaRed, StatusYIKRed, StatusReddedildi:
		return true
	default:
		return false
	}
}

// IsApproved checks if the application was approved
func (a *Application) IsApproved() bool {
	return a.Status == StatusKabul
}

// IsOnursal checks if this is an honorary membership application
func (a *Application) IsOnursal() bool {
	return a.MembershipType == MembershipOnursal
}

// RequiresReferences checks if this membership type requires reference verification
func (a *Application) RequiresReferences() bool {
	return a.MembershipType == MembershipAsil || a.MembershipType == MembershipAkademik
}

// RequiresConsultation checks if this membership type requires consultation
func (a *Application) RequiresConsultation() bool {
	return a.MembershipType == MembershipProfesyonel || a.MembershipType == MembershipOgrenci
}
