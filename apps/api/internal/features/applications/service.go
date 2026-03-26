package applications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/shared"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service handles application business logic.
type Service struct {
	repo     *Repository
	authRepo *auth.Repository
	logRepo  *logs.Repository
}

// NewService creates a new application service.
func NewService(repo *Repository, authRepo *auth.Repository, logRepo *logs.Repository) *Service {
	return &Service{
		repo:     repo,
		authRepo: authRepo,
		logRepo:  logRepo,
	}
}

// Submit validates and creates a new application.
func (s *Service) Submit(ctx context.Context, req *CreateApplicationRequest, actorID string) (*SubmitResult, error) {
	// 1. Validate membership type business rules
	if err := s.validateBusinessRules(req); err != nil {
		return nil, err
	}

	// 2. Check LinkedIn URL uniqueness (required for asil/akademik)
	if req.LinkedInURL != "" {
		unique, err := s.repo.CheckLinkedInUniqueness(ctx, req.LinkedInURL, "")
		if err != nil {
			return nil, fmt.Errorf("applications: check linkedin: %w", err)
		}
		if !unique {
			return nil, fmt.Errorf("linkedin_url: %w", errors.New("a application with this LinkedIn URL already exists"))
		}
	}

	// 3. Build application entity
	app := &Application{
		ApplicantName:  req.ApplicantName,
		ApplicantEmail: req.ApplicantEmail,
		ApplicantPhone: req.ApplicantPhone,
		LinkedInURL:    req.LinkedInURL,
		PhotoURL:       req.PhotoURL,
		MembershipType: req.MembershipType,
		ProposalReason: req.ProposalReason,
		Status:         GetInitialStatus(req.MembershipType),
	}
	if req.ProposedByUserID != "" {
		app.ProposedByUserID = &req.ProposedByUserID
	}

	if err := s.repo.Create(ctx, app); err != nil {
		return nil, fmt.Errorf("applications: create: %w", err)
	}

	// 4. Log creation
	metadata, _ := json.Marshal(map[string]interface{}{
		"applicant_email": app.ApplicantEmail,
		"membership_type": app.MembershipType,
		"initial_status":  app.Status,
	})
	var actorIDPtr *string
	if actorID != "" {
		actorIDPtr = &actorID
	}
	_ = s.logRepo.Create(ctx, &logs.Log{
		Action:     "application.created",
		ActorID:    actorIDPtr,
		EntityType: "application",
		EntityID:   app.ID,
		Metadata:   datatypes.JSON(metadata),
	})

	// 5. Check for repeat applicant (query terminated logs for same email)
	result := &SubmitResult{Application: app}
	prevAppID, err := s.findTerminatedByEmail(ctx, req.ApplicantEmail)
	if err == nil && prevAppID != "" {
		result.RepeatApplicant = true
		result.PreviousAppID = &prevAppID
		// Store previous_app_id on the application row
		_ = s.repo.db.WithContext(ctx).
			Table("applications").
			Where("id = ?", app.ID).
			Update("previous_app_id", prevAppID)
	}

	return result, nil
}

// GetByID returns the full application detail, masking sensitive fields based on role.
func (s *Service) GetByID(ctx context.Context, id, requestorRole string) (*ApplicationDetailResponse, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("applications: get by id: %w", err)
	}

	resp := &ApplicationDetailResponse{
		ID:                  app.ID,
		ApplicantName:       app.ApplicantName,
		ApplicantEmail:      app.ApplicantEmail,
		ApplicantPhone:      app.ApplicantPhone,
		LinkedInURL:         app.LinkedInURL,
		PhotoURL:            app.PhotoURL,
		MembershipType:      app.MembershipType,
		Status:              app.Status,
		ProposalReason:      app.ProposalReason,
		PreviousAppID:       app.PreviousAppID,
		CreatedAt:           app.CreatedAt,
		UpdatedAt:           app.UpdatedAt,
		AllowedNextStatuses: AllowedTransitions(app.MembershipType, app.Status),
	}

	// Rejection info only visible to yk / admin
	isPrivileged := requestorRole == "yk" || requestorRole == "admin"
	if isPrivileged {
		resp.RejectionReason = app.RejectionReason
		resp.RejectedByRole = app.RejectedByRole
	}

	// Repeat applicant flag
	if isPrivileged && app.PreviousAppID != nil {
		resp.RepeatApplicant = true
	}

	return resp, nil
}

// ListAll returns a paginated, filtered list of applications.
func (s *Service) ListAll(ctx context.Context, filters ApplicationFilters) (*ApplicationListResponse, error) {
	apps, total, err := s.repo.FindAll(ctx, filters)
	if err != nil {
		return nil, err
	}

	page := filters.Page
	if page < 1 {
		page = 1
	}
	pageSize := filters.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	summaries := make([]ApplicationSummary, 0, len(apps))
	for _, a := range apps {
		summaries = append(summaries, ApplicationSummary{
			ID:             a.ID,
			ApplicantName:  a.ApplicantName,
			ApplicantEmail: a.ApplicantEmail,
			MembershipType: a.MembershipType,
			Status:         a.Status,
			CreatedAt:      a.CreatedAt,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &ApplicationListResponse{
		Data:       summaries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTimeline returns the status-change history for an application.
func (s *Service) GetTimeline(ctx context.Context, id string) ([]TimelineEntry, error) {
	// Verify application exists
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return s.repo.GetTimeline(ctx, id)
}

// GetRedHistory returns all terminated applications from the same email.
// Only allowed for yk / admin roles.
func (s *Service) GetRedHistory(ctx context.Context, id string) ([]*Application, error) {
	app, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}

	all, err := s.repo.FindByApplicantEmail(ctx, app.ApplicantEmail)
	if err != nil {
		return nil, err
	}

	// Filter: only previously terminated applications (not the current one)
	var history []*Application
	for _, a := range all {
		if a.ID != id && IsTerminated(a.Status) {
			history = append(history, a)
		}
	}
	return history, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// validateBusinessRules enforces membership-type-specific requirements.
func (s *Service) validateBusinessRules(req *CreateApplicationRequest) error {
	switch req.MembershipType {
	case MembershipAsil, MembershipAkademik:
		if req.LinkedInURL == "" {
			return fmt.Errorf("linkedin_url: required for %s membership", req.MembershipType)
		}
		if len(req.References) < 3 {
			return fmt.Errorf("references: at least 3 references required for %s membership", req.MembershipType)
		}
	case MembershipOnursal:
		if len(req.ProposalReason) < 100 {
			return fmt.Errorf("proposal_reason: must be at least 100 characters for onursal membership")
		}
		if req.ProposedByUserID == "" {
			return fmt.Errorf("proposed_by_user_id: required for onursal membership")
		}
	}
	return nil
}

// findTerminatedByEmail queries audit logs for the most recent terminated application
// by the given email address. Returns the application ID or empty string.
func (s *Service) findTerminatedByEmail(ctx context.Context, email string) (string, error) {
	type result struct {
		EntityID string `gorm:"column:entity_id"`
	}
	var r result
	err := s.logRepo.DB().WithContext(ctx).Raw(`
		SELECT entity_id
		FROM logs
		WHERE entity_type = 'application'
		  AND action = 'application.terminated'
		  AND JSON_UNQUOTE(JSON_EXTRACT(metadata, '$.applicant_email')) = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, email).Scan(&r).Error
	if err != nil {
		return "", err
	}
	return r.EntityID, nil
}
