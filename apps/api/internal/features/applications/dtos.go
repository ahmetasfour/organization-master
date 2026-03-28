package applications

import (
	"membership-system/api/internal/shared"
	"time"
)

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// ReferenceInput is aliased from shared package to avoid circular imports.
type ReferenceInput = shared.ReferenceInput

// CreateApplicationRequest is the payload for POST /api/v1/applications.
type CreateApplicationRequest struct {
	ApplicantName    string           `json:"applicant_name"  validate:"required,min=2,max=255"`
	ApplicantEmail   string           `json:"applicant_email" validate:"required,email"`
	ApplicantPhone   string           `json:"applicant_phone" validate:"omitempty,e164"` // E.164 format
	LinkedInURL      string           `json:"linkedin_url"    validate:"omitempty,linkedin_url"`
	PhotoURL         string           `json:"photo_url"       validate:"omitempty,photo_url"`
	MembershipType   MembershipType   `json:"membership_type" validate:"required,oneof=asil akademik profesyonel öğrenci onursal"`
	ProposalReason   string           `json:"proposal_reason"`
	ProposedByUserID string           `json:"proposed_by_user_id"`
	References       []ReferenceInput `json:"references"`
}

// AdvanceStatusRequest is the payload for PATCH /api/v1/applications/:id/advance.
type AdvanceStatusRequest struct {
	TargetStatus ApplicationStatus `json:"target_status" validate:"required"`
}

// ─── Response DTOs ─────────────────────────────────────────────────────────────

// TimelineEntry represents a single status-change event in an application's history.
type TimelineEntry struct {
	Status    ApplicationStatus `json:"status"`
	ChangedAt time.Time         `json:"changed_at"`
	ChangedBy string            `json:"changed_by,omitempty"` // actor email or role
	Notes     string            `json:"notes,omitempty"`
}

// ApplicationSummary is used in list responses.
type ApplicationSummary struct {
	ID             string            `json:"id"`
	ApplicantName  string            `json:"applicant_name"`
	ApplicantEmail string            `json:"applicant_email"`
	MembershipType MembershipType    `json:"membership_type"`
	Status         ApplicationStatus `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
}

// ApplicationDetailResponse is the full application payload returned on GET /:id.
type ApplicationDetailResponse struct {
	ID             string            `json:"id"`
	ApplicantName  string            `json:"applicant_name"`
	ApplicantEmail string            `json:"applicant_email"`
	ApplicantPhone string            `json:"applicant_phone,omitempty"`
	LinkedInURL    string            `json:"linkedin_url,omitempty"`
	PhotoURL       string            `json:"photo_url,omitempty"`
	MembershipType MembershipType    `json:"membership_type"`
	Status         ApplicationStatus `json:"status"`
	ProposalReason string            `json:"proposal_reason,omitempty"`

	// Rejection info — only exposed to yk / admin
	RejectionReason *string `json:"rejection_reason,omitempty"`
	RejectedByRole  string  `json:"rejected_by_role,omitempty"`

	// Repeat applicant flag — only set for yk / admin
	RepeatApplicant bool    `json:"repeat_applicant,omitempty"`
	PreviousAppID   *string `json:"previous_app_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AllowedNextStatuses []ApplicationStatus `json:"allowed_next_statuses"`
}

// ApplicationListResponse wraps a paginated list.
type ApplicationListResponse struct {
	Data       []ApplicationSummary `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// ApplicationFilters holds query parameters for filtering application lists.
type ApplicationFilters struct {
	MembershipType string `query:"membership_type"`
	Status         string `query:"status"`
	Search         string `query:"search"`
	Page           int    `query:"page"`
	PageSize       int    `query:"page_size"`
}

// SubmitResult wraps the created application plus any repeat-applicant metadata.
type SubmitResult struct {
	Application     *Application `json:"application"`
	RepeatApplicant bool         `json:"repeat_applicant"`
	PreviousAppID   *string      `json:"previous_app_id,omitempty"`
}

// LinkedInField is a helper field name used for validation
const LinkedInField = "linkedin_url"
