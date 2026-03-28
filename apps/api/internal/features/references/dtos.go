package references

// ─── Request DTOs ──────────────────────────────────────────────────────────────

// ReferenceResponseRequest is the payload for POST /api/v1/ref/respond/:token.
type ReferenceResponseRequest struct {
	ResponseType string `json:"response_type" validate:"required,oneof=positive unknown negative"`
	// Reason is required only when response_type = negative (min 30 chars).
	Reason string `json:"reason"`
}

// SubmitReplacementRequest is the payload for POST /api/v1/ref/replace/:token.
type SubmitReplacementRequest struct {
	RefereeName  string `json:"referee_name" validate:"required,min=2"`
	RefereeEmail string `json:"referee_email" validate:"required,email"`
}

// ─── Response DTOs ─────────────────────────────────────────────────────────────

// ReferenceFormData is returned on GET /api/v1/ref/respond/:token.
// Contains just enough context for the public form — no sensitive application data.
type ReferenceFormData struct {
	ApplicantName  string `json:"applicant_name"`
	MembershipType string `json:"membership_type"`
	RefereeName    string `json:"referee_name"`
	ExpiresAt      string `json:"expires_at"` // formatted date string
}

// ReplacementFormData is returned on GET /api/v1/ref/replace/:token.
type ReplacementFormData struct {
	ApplicantName      string `json:"applicant_name"`
	MembershipType     string `json:"membership_type"`
	UnknownRefereeName string `json:"unknown_referee_name"`
	ApplicationID      string `json:"application_id"`
}

// ReferenceSummary is used in admin views.
type ReferenceSummary struct {
	ID            string  `json:"id"`
	ApplicationID string  `json:"application_id"`
	RefereeName   string  `json:"referee_name"`
	RefereeEmail  string  `json:"referee_email"`
	IsReplacement bool    `json:"is_replacement"`
	Round         int     `json:"round"`
	ResponseType  *string `json:"response_type,omitempty"`
	Status        string  `json:"status"` // pending | positive | negative | unknown | expired
}

// ─── Cross-package helpers ─────────────────────────────────────────────────────

// AppContext carries the minimal application data needed by reference operations.
// Defined here to avoid circular imports with the applications package.
type AppContext struct {
	ID             string
	ApplicantName  string
	ApplicantEmail string
	MembershipType string
}
