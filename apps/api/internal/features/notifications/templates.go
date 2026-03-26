package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"
)

// ─── Template data structs ──────────────────────────────────────────────────────

// ReferenceRequestData is the data passed to reference_request.html.
type ReferenceRequestData struct {
	RefereeName    string
	ApplicantName  string
	MembershipType string
	ResponseURL    string
	ExpiresAt      string
}

// ReferenceReminderData is the data passed to reference_reminder.html.
type ReferenceReminderData struct {
	RefereeName    string
	ApplicantName  string
	ResponseURL    string
	HoursRemaining int
}

// NewRefNeededData is the data passed to applicant_new_ref_needed.html.
type NewRefNeededData struct {
	ApplicantName      string
	UnknownRefereeName string
	PortalURL          string
}

// ReputationQueryData is the data passed to reputation_query.html.
type ReputationQueryData struct {
	ContactName   string
	ApplicantName string
	ApplicantURL  string
	ResponseURL   string
	ExpiresAt     string
}

// ConsultationData is the data passed to consultation_request.html.
type ConsultationData struct {
	MemberName        string
	ApplicantName     string
	ApplicantLinkedIn string
	MembershipType    string
	ResponseURL       string
	ExpiresAt         string
}

// AcceptedData is the data passed to application_accepted.html.
type AcceptedData struct {
	ApplicantName  string
	MembershipType string
}

// RejectedData is the data passed to application_rejected.html.
// NOTE: rejection_reason is intentionally omitted — never disclose why.
type RejectedData struct {
	ApplicantName string
}

// HonoraryProposalData is the data passed to honorary_proposal_notify.html.
// Sent to each YK member when an honorary membership proposal is submitted.
type HonoraryProposalData struct {
	YKMemberName    string
	ProposerName    string
	NomineeName     string
	NomineeLinkedIn string
	ProposalReason  string
	ReviewURL       string
}

// ─── Renderer ──────────────────────────────────────────────────────────────────

// Render parses an HTML template string and executes it with the provided data.
// Returns (htmlBody, textBody, error).
func Render(tmplContent string, data interface{}) (string, string, error) {
	tmpl, err := template.New("email").Parse(tmplContent)
	if err != nil {
		return "", "", fmt.Errorf("notifications: parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", "", fmt.Errorf("notifications: render template: %w", err)
	}

	htmlBody := buf.String()
	textBody := stripHTML(htmlBody)
	return htmlBody, textBody, nil
}

// FormatTime formats a time.Time for display in email templates.
func FormatTime(t time.Time) string {
	return t.Format("02 Jan 2006 15:04 MST")
}

// stripHTML removes HTML tags to produce a plain-text version.
func stripHTML(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	// Collapse multiple newlines
	text := strings.TrimSpace(result.String())
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return text
}
