package shared

// ReferenceInput represents a referee contact for application submissions.
// Used to avoid circular dependencies between applications and references packages.
type ReferenceInput struct {
	RefereeName  string `json:"referee_name"  validate:"required,min=2,max=255"`
	RefereeEmail string `json:"referee_email" validate:"required,email"`
}
