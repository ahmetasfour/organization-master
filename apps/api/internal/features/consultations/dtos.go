package consultations

// ─── Request DTOs ──────────────────────────────────────────────────────────────

// AddConsultationsRequest is the payload for POST /api/v1/applications/:id/consultations.
type AddConsultationsRequest struct {
	Consultees []ConsulteeInput `json:"consultees" validate:"required,min=2,dive"`
}

// ConsulteeInput identifies a single member to be consulted.
type ConsulteeInput struct {
	UserID string `json:"user_id" validate:"required,uuid4"`
}

// ConsultationResponseRequest is the payload for POST /api/v1/consult/respond/:token.
type ConsultationResponseRequest struct {
	ResponseType string `json:"response_type" validate:"required,oneof=positive negative"`
	// Reason is required when response_type = negative (min 30 chars).
	Reason string `json:"reason"`
}

// ─── Response DTOs ─────────────────────────────────────────────────────────────

// ConsultationFormData is returned on GET /api/v1/consult/respond/:token.
// Contains just enough context for the public form — no sensitive application data.
type ConsultationFormData struct {
	ApplicantName  string `json:"applicant_name"`
	MembershipType string `json:"membership_type"`
	MemberName     string `json:"member_name"`
	ExpiresAt      string `json:"expires_at"` // human-readable date string
}

// ConsultationSummary is used in admin views.
type ConsultationSummary struct {
	ID            string  `json:"id"`
	ApplicationID string  `json:"application_id"`
	MemberUserID  string  `json:"member_user_id"`
	MemberName    string  `json:"member_name"`
	MemberEmail   string  `json:"member_email"`
	ResponseType  *string `json:"response_type,omitempty"`
	Status        string  `json:"status"` // pending | positive | negative | expired
	CreatedAt     string  `json:"created_at"`
}
