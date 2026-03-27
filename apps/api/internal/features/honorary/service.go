package honorary

import (
	"context"
	"fmt"

	"membership-system/api/internal/features/applications"
	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/notifications"

	"github.com/google/uuid"
)

type Service struct {
	repo             *Repository
	authRepo         *auth.Repository
	notificationsSvc *notifications.Service
}

func NewService(
	repo *Repository,
	authRepo *auth.Repository,
	notificationsSvc *notifications.Service,
) *Service {
	return &Service{
		repo:             repo,
		authRepo:         authRepo,
		notificationsSvc: notificationsSvc,
	}
}

func (s *Service) Propose(ctx context.Context, req ProposeRequest, proposerID string) (*applications.Application, error) {
	// 1. Validate proposer role
	proposer, err := s.authRepo.FindByID(ctx, proposerID)
	if err != nil {
		return nil, fmt.Errorf("failed to find proposer: %w", err)
	}

	if proposer.Role != auth.RoleAsilUye && proposer.Role != auth.RoleYIKUye {
		return nil, fmt.Errorf("only asil_uye and yik_uye can propose honorary members")
	}

	// 2. Check LinkedIn URL uniqueness
	exists, err := s.repo.CheckLinkedInExists(ctx, req.NomineeLinkedIn)
	if err != nil {
		return nil, fmt.Errorf("failed to check LinkedIn uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("LinkedIn URL already exists in the system")
	}

	// 3. Create application
	applicationID := uuid.New().String()
	application := &applications.Application{
		ID:               applicationID,
		ApplicantName:    req.NomineeName,
		ApplicantEmail:   "nominee+" + uuid.New().String()[:8] + "@honorary.placeholder", // Unique placeholder for honorary proposals
		LinkedInURL:      req.NomineeLinkedIn,
		MembershipType:   "onursal",
		Status:           "öneri_alındı",
		ProposedByUserID: &proposerID,
		ProposalReason:   req.ProposalReason,
	}

	err = s.repo.Create(ctx, application)
	if err != nil {
		return nil, fmt.Errorf("failed to create honorary proposal: %w", err)
	}

	// 4. Load YK members
	ykMembers, err := s.repo.GetYKMembers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get YK members: %w", err)
	}

	// 5. Send notifications to YK members
	proposerName := proposer.FullName

	// Convert to expected struct format
	ykMembersList := make([]struct {
		ID    string
		Email string
		Name  string
	}, len(ykMembers))

	for i, member := range ykMembers {
		ykMembersList[i] = struct {
			ID    string
			Email string
			Name  string
		}{
			ID:    member.ID,
			Email: member.Email,
			Name:  member.FullName,
		}
	}

	err = s.notificationsSvc.SendHonoraryProposal(
		ctx,
		applicationID,
		proposerName,
		req.NomineeName,
		req.NomineeLinkedIn,
		req.ProposalReason,
		ykMembersList,
	)
	if err != nil {
		// Log error but don't fail the proposal creation
		fmt.Printf("Warning: Failed to send notification emails: %v\n", err)
	}

	return application, nil
}

func (s *Service) ListProposals(ctx context.Context) ([]*ProposalResponse, error) {
	return s.repo.FindAll(ctx)
}
