package reputation

import "time"

// ─── Inbound DTOs ──────────────────────────────────────────────────────────────

// ContactInput represents a single contact in the batch-add request.
type ContactInput struct {
	Name  string `json:"name"  validate:"required,min=2,max=255"`
	Email string `json:"email" validate:"required,email"`
}

// AddContactsRequest is the payload for adding reputation contacts.
// CRITICAL: Exactly 10 contacts are required by spec.
type AddContactsRequest struct {
	Contacts []ContactInput `json:"contacts" validate:"required,len=10,dive"`
}

// ContactResponseRequest is the payload submitted by a contact through the public page.
type ContactResponseRequest struct {
	ResponseType string `json:"response_type" validate:"required,oneof=clean negative"`
	// Reason is required when ResponseType = "negative", min 30 chars.
	Reason string `json:"reason" validate:"required_if=ResponseType negative,omitempty,min=30"`
}

// ─── Outbound DTOs ─────────────────────────────────────────────────────────────

// ReputationFormData is returned by the public GET token endpoint so the contact
// can see context before submitting their response.
type ReputationFormData struct {
	ContactName       string `json:"contact_name"`
	ApplicantName     string `json:"applicant_name"`
	ApplicantLinkedIn string `json:"applicant_linkedin"`
	ExpiresAt         string `json:"expires_at"`
}

// ContactStatus is one row inside ReputationStatusResponse.
// Contact email is always masked: j***@example.com
type ContactStatus struct {
	ID          string     `json:"id"`
	ContactName string     `json:"contact_name"`
	Email       string     `json:"email"`  // masked
	Status      string     `json:"status"` // pending | clean | flagged
	RespondedAt *time.Time `json:"responded_at,omitempty"`
}

// ReputationStatusResponse is returned by GET /applications/:id/reputation.
type ReputationStatusResponse struct {
	ApplicationID string          `json:"application_id"`
	TotalContacts int             `json:"total_contacts"`
	Responded     int             `json:"responded"`
	Clean         int             `json:"clean"`
	Flagged       int             `json:"flagged"`
	Contacts      []ContactStatus `json:"contacts"`
}
