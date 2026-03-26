package references

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResponseType represents the referee's response classification
type ResponseType string

const (
	ResponsePositive ResponseType = "positive"
	ResponseUnknown  ResponseType = "unknown"
	ResponseNegative ResponseType = "negative"
)

// Reference represents a reference request sent to a referee
type Reference struct {
	ID                  string `gorm:"type:char(36);primaryKey" json:"id"`
	ApplicationID       string `gorm:"type:char(36);not null;index" json:"application_id"`
	RefereeName         string `gorm:"type:varchar(255);not null" json:"referee_name"`
	RefereeEmail        string `gorm:"type:varchar(255);not null;index" json:"referee_email"`
	RefereePhone        string `gorm:"type:varchar(50)" json:"referee_phone,omitempty"`
	RefereeOrganization string `gorm:"type:varchar(255)" json:"referee_organization,omitempty"`

	// Token security: Store SHA-256 hash only, never the raw token
	TokenHash      string     `gorm:"type:varchar(64);not null;index" json:"-"`
	TokenExpiresAt time.Time  `gorm:"not null;index" json:"token_expires_at"`
	TokenUsedAt    *time.Time `gorm:"type:timestamp null" json:"token_used_at,omitempty"`

	// Reference tracking
	IsReplacement bool `gorm:"default:false;not null" json:"is_replacement"`
	Round         int  `gorm:"default:1;not null" json:"round"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;autoUpdateTime" json:"updated_at"`

	// Associations
	Response *ReferenceResponse `gorm:"foreignKey:ReferenceID" json:"response,omitempty"`
}

// TableName specifies the table name for GORM
func (Reference) TableName() string {
	return "references"
}

// BeforeCreate GORM hook to generate UUID before creating a reference
func (r *Reference) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	return nil
}

// IsTokenExpired checks if the reference token has expired
func (r *Reference) IsTokenExpired() bool {
	return time.Now().After(r.TokenExpiresAt)
}

// IsTokenUsed checks if the reference token has been used
func (r *Reference) IsTokenUsed() bool {
	return r.TokenUsedAt != nil
}

// IsSubmitted checks if the referee has submitted a response
func (r *Reference) IsSubmitted() bool {
	return r.Response != nil
}

// ReferenceResponse represents a referee's submitted response
type ReferenceResponse struct {
	ID                      string       `gorm:"type:char(36);primaryKey" json:"id"`
	ReferenceID             string       `gorm:"type:char(36);not null;uniqueIndex:unique_response_per_reference" json:"reference_id"`
	ResponseType            ResponseType `gorm:"type:enum('positive','unknown','negative');not null;index" json:"response_type"`
	RelationshipDescription string       `gorm:"type:text" json:"relationship_description,omitempty"`
	DurationYears           int          `gorm:"type:int" json:"duration_years,omitempty"`
	RecommendationText      string       `gorm:"type:text;not null" json:"recommendation_text"`
	CreatedAt               time.Time    `gorm:"not null;autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for GORM
func (ReferenceResponse) TableName() string {
	return "reference_responses"
}

// BeforeCreate GORM hook to generate UUID before creating a response
func (rr *ReferenceResponse) BeforeCreate(tx *gorm.DB) error {
	if rr.ID == "" {
		rr.ID = uuid.New().String()
	}
	return nil
}

// IsPositive checks if the response is positive
func (rr *ReferenceResponse) IsPositive() bool {
	return rr.ResponseType == ResponsePositive
}

// IsNegative checks if the response is negative
func (rr *ReferenceResponse) IsNegative() bool {
	return rr.ResponseType == ResponseNegative
}
