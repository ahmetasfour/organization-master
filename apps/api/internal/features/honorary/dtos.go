package honorary

import "time"

type ProposeRequest struct {
	NomineeName     string `json:"nominee_name" validate:"required,min=2"`
	NomineeLinkedIn string `json:"nominee_linkedin" validate:"required,url"`
	ProposalReason  string `json:"proposal_reason" validate:"required,min=100"`
}

type ProposalResponse struct {
	ApplicationID   string    `json:"application_id"`
	NomineeName     string    `json:"nominee_name"`
	NomineeLinkedIn string    `json:"nominee_linkedin"`
	ProposalReason  string    `json:"proposal_reason"`
	ProposedBy      string    `json:"proposed_by"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}
