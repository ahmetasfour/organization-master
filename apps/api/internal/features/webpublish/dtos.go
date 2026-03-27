package webpublish

import "time"

// RecordConsentRequest represents the request to record web publish consent
type RecordConsentRequest struct {
	Consented bool `json:"consented"`
}

// MemberListItem represents a published member in the public list
type MemberListItem struct {
	FullName       string `json:"full_name"`
	MembershipType string `json:"membership_type"`
	AcceptedAt     string `json:"accepted_at"`
}

// ConsentResponse represents the response after recording consent
type ConsentResponse struct {
	ApplicationID string `json:"application_id"`
	Consented     bool   `json:"consented"`
	IsPublished   bool   `json:"is_published"`
	RecordedAt    string `json:"recorded_at"`
}

// WebPublishConsent represents the consent record
type WebPublishConsent struct {
	ID            string    `json:"id" gorm:"type:char(36);primaryKey"`
	ApplicationID string    `json:"application_id" gorm:"type:char(36);uniqueIndex;not null"`
	Consented     bool      `json:"consented" gorm:"not null"`
	RecordedBy    string    `json:"recorded_by" gorm:"column:recorded_by;type:char(36)"`
	CreatedAt     time.Time `json:"created_at"`
}

func (WebPublishConsent) TableName() string {
	return "web_publish_consents"
}
