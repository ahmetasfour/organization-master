package voting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"membership-system/api/internal/features/auth"
	"membership-system/api/internal/features/logs"
	"membership-system/api/internal/shared"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// voteAppRow is a minimal projection of applications table.
// Inline to avoid a circular import with the applications package.
type voteAppRow struct {
	ID             string `gorm:"column:id"`
	MembershipType string `gorm:"column:membership_type"`
	Status         string `gorm:"column:status"`
	ApplicantName  string `gorm:"column:applicant_name"`
	ApplicantEmail string `gorm:"column:applicant_email"`
}

func (voteAppRow) TableName() string { return "applications" }

// voteUserRow is a minimal projection of the users table for voter name lookups.
type voteUserRow struct {
	ID       string `gorm:"column:id"`
	FullName string `gorm:"column:full_name"`
	Role     string `gorm:"column:role"`
}

func (voteUserRow) TableName() string { return "users" }

// ─── Stage / Status maps ──────────────────────────────────────────────────────

// validStagesForType maps each membership type to the vote stages it must pass.
var validStagesForType = map[string]map[VoteStage]bool{
	"asil":        {VoteStageYKPrelim: true, VoteStageYKFinal: true},
	"akademik":    {VoteStageYKPrelim: true, VoteStageYKFinal: true},
	"profesyonel": {VoteStageYKFinal: true},
	"öğrenci":     {VoteStageYKFinal: true},
	"onursal":     {VoteStageYKPrelim: true, VoteStageYIK: true, VoteStageYKFinal: true},
}

// stageRequiredStatus maps a VoteStage to the application status that must be
// present before that stage's voting can begin.
var stageRequiredStatus = map[VoteStage]string{
	VoteStageYKPrelim: "yk_ön_incelemede",
	VoteStageYIK:      "yik_değerlendirmede",
	VoteStageYKFinal:  "gündemde",
}

// stageAdvancesTo maps a VoteStage to the next status on unanimous approval.
var stageAdvancesTo = map[VoteStage]string{
	VoteStageYKPrelim: "ön_onaylandı",
	VoteStageYIK:      "gündemde",
	VoteStageYKFinal:  "kabul",
}

// stageVoterRole maps each stage to the role of eligible voters.
var stageVoterRole = map[VoteStage]string{
	VoteStageYKPrelim: "yk",
	VoteStageYIK:      "yik",
	VoteStageYKFinal:  "yk",
}

// ─── Service ──────────────────────────────────────────────────────────────────

// Service contains business logic for the three-stage voting system.
type Service struct {
	repo     *Repository
	authRepo *auth.Repository
	logRepo  *logs.Repository
	db       *gorm.DB
}

// NewService creates a new voting service.
func NewService(
	repo *Repository,
	authRepo *auth.Repository,
	logRepo *logs.Repository,
	db *gorm.DB,
) *Service {
	return &Service{
		repo:     repo,
		authRepo: authRepo,
		logRepo:  logRepo,
		db:       db,
	}
}

// ─── CastVote ─────────────────────────────────────────────────────────────────

// CastVote validates and records a vote for the given application at the given
// stage, enforcing:
//   - application must not be terminated (RedGuard)
//   - stage must be valid for the application's membership type
//   - application must be in the required status for the stage
//   - voterRole must match the stage's eligible role
//   - one vote per (application, voter, stage) — ErrDuplicateVote on collision
//   - reject votes are treated as vetoes: RedGuard.Terminate immediately
//   - when all eligible voters have approved (no reject): advance status
func (s *Service) CastVote(
	ctx context.Context,
	appID, voterID, voterRole string,
	stage VoteStage,
	req *CastVoteRequest,
) error {
	// 1. RedGuard: ensure application is not already terminated
	redGuard := shared.NewRedGuard(s.db)
	if err := redGuard.AssertNotTerminated(ctx, appID); err != nil {
		return err
	}

	// 2. Load application
	var app voteAppRow
	if err := s.db.WithContext(ctx).First(&app, "id = ?", appID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return shared.ErrNotFound
		}
		return fmt.Errorf("voting: load application: %w", err)
	}

	// 3. Validate stage is applicable for this membership type
	allowedStages, ok := validStagesForType[app.MembershipType]
	if !ok || !allowedStages[stage] {
		return fmt.Errorf("%w: stage %q does not apply to membership type %q",
			shared.ErrForbidden, stage, app.MembershipType)
	}

	// 4. Validate application is in the required status for this stage
	requiredStatus, _ := stageRequiredStatus[stage]
	if app.Status != requiredStatus {
		return fmt.Errorf("%w: application must be in status %q to vote at stage %q, got %q",
			shared.ErrForbidden, requiredStatus, stage, app.Status)
	}

	// 5. Validate voter role matches the stage
	eligibleRole := stageVoterRole[stage]
	if voterRole != eligibleRole {
		return fmt.Errorf("%w: stage %q requires role %q, caller has role %q",
			shared.ErrForbidden, stage, eligibleRole, voterRole)
	}

	// 6. Check for duplicate vote
	existing, err := s.repo.FindByApplicationVoterStage(ctx, appID, voterID, stage)
	if err != nil {
		return fmt.Errorf("voting: duplicate check: %w", err)
	}
	if existing != nil {
		return shared.ErrDuplicateVote
	}

	// 7. Validate reject reason length
	if req.VoteType == string(VoteReject) && len([]rune(req.Reason)) < 20 {
		return fmt.Errorf("reason: minimum 20 characters required when rejecting")
	}

	// 8. Persist vote
	isVeto := req.VoteType == string(VoteReject)
	vote := &Vote{
		ApplicationID: appID,
		VoterID:       voterID,
		VoteStage:     stage,
		VoteType:      VoteType(req.VoteType),
		IsVeto:        isVeto,
		Reason:        req.Reason,
	}
	if err := s.repo.Create(ctx, vote); err != nil {
		return fmt.Errorf("voting: persist vote: %w", err)
	}

	_ = s.writeLog(ctx, "vote.cast", appID, "application", map[string]interface{}{
		"voter_id":  voterID,
		"role":      voterRole,
		"stage":     string(stage),
		"vote_type": req.VoteType,
		"is_veto":   isVeto,
	})

	// 9. Veto path: any single reject terminates the application immediately
	if isVeto {
		reason := fmt.Sprintf("Üye adayı %s oylamasında red oyu aldı. Gerekçe: %s", string(stage), req.Reason)
		if termErr := redGuard.Terminate(ctx, appID, reason, voterID, voterRole); termErr != nil {
			// Already terminated by a concurrent vote — not a hard error
			if errors.Is(termErr, shared.ErrApplicationTerminated) {
				return nil
			}
			return fmt.Errorf("voting: terminate on veto: %w", termErr)
		}
		_ = s.writeLog(ctx, "vote.veto", appID, "application", map[string]interface{}{
			"voter_id": voterID,
			"role":     voterRole,
			"stage":    string(stage),
			"reason":   req.Reason,
		})
		return nil
	}

	// 10. Completion check: if every eligible voter has voted (no reject) → advance
	if err := s.tryAdvanceStatus(ctx, appID, voterRole, stage, voterID); err != nil {
		// Log but don't surface — vote was already saved
		_ = s.writeLog(ctx, "vote.advance_failed", appID, "application", map[string]interface{}{
			"error": err.Error(),
			"stage": string(stage),
		})
	}

	return nil
}

// tryAdvanceStatus checks whether all active voters for the stage have voted.
// If so, it advances the application status to the next stage's starting point.
func (s *Service) tryAdvanceStatus(
	ctx context.Context,
	appID, voterRole string,
	stage VoteStage,
	actorID string,
) error {
	// Count total eligible voters
	totalVoters, err := s.repo.CountActiveVotersByRole(ctx, voterRole)
	if err != nil {
		return fmt.Errorf("voting: count voters: %w", err)
	}
	if totalVoters == 0 {
		return nil // Nothing to check
	}

	// Count how many approved or abstained in this stage
	approved, err := s.repo.CountVotesByStageAndType(ctx, appID, stage, VoteApprove)
	if err != nil {
		return fmt.Errorf("voting: count approvals: %w", err)
	}
	abstained, err := s.repo.CountVotesByStageAndType(ctx, appID, stage, VoteAbstain)
	if err != nil {
		return fmt.Errorf("voting: count abstentions: %w", err)
	}

	votesIn := approved + abstained
	if votesIn < totalVoters {
		// Not everyone has voted yet
		return nil
	}

	// All votes cast and no reject (reject would have terminated already) → advance
	nextStatus := stageAdvancesTo[stage]
	prevStatus := stageRequiredStatus[stage]

	if err := s.db.WithContext(ctx).
		Exec("UPDATE applications SET status = ?, updated_at = ? WHERE id = ? AND status = ?",
			nextStatus, time.Now(), appID, prevStatus).Error; err != nil {
		return fmt.Errorf("voting: advance status to %s: %w", nextStatus, err)
	}

	_ = s.writeLog(ctx, "status.change", appID, "application", map[string]interface{}{
		"from":   prevStatus,
		"to":     nextStatus,
		"actor":  actorID,
		"reason": fmt.Sprintf("all %d eligible %s voters cast votes (stage: %s)", totalVoters, voterRole, stage),
	})

	return nil
}

// ─── GetVotes ─────────────────────────────────────────────────────────────────

// GetVotes returns a summary of votes for the given application+stage.
// If requestorRole is "yk" or "admin", individual vote details (including
// voter names and rejection reasons) are included. Others only see aggregates.
func (s *Service) GetVotes(
	ctx context.Context,
	appID string,
	stage VoteStage,
	requestorRole string,
) (*VoteSummaryResponse, error) {
	// Load application to check termination status
	var app voteAppRow
	if err := s.db.WithContext(ctx).First(&app, "id = ?", appID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("voting: load application: %w", err)
	}

	votes, err := s.repo.FindByApplicationAndStage(ctx, appID, stage)
	if err != nil {
		return nil, fmt.Errorf("voting: find votes: %w", err)
	}

	eligibleRole := stageVoterRole[stage]
	totalVoters, err := s.repo.CountActiveVotersByRole(ctx, eligibleRole)
	if err != nil {
		return nil, fmt.Errorf("voting: count voters: %w", err)
	}

	canSeeDetails := requestorRole == "yk" || requestorRole == "admin"

	// Determine termination flag
	isTerminated := isTerminalStatus(app.Status)

	var approved, abstained, rejected int64
	voteResponses := make([]VoteResponse, 0, len(votes))

	for _, v := range votes {
		switch v.VoteType {
		case VoteApprove:
			approved++
		case VoteAbstain:
			abstained++
		case VoteReject:
			rejected++
		}

		if canSeeDetails {
			// Resolve voter name
			voterName := s.resolveVoterName(ctx, v.VoterID)

			var reasonPtr *string
			if v.Reason != "" {
				r := v.Reason
				reasonPtr = &r
			}

			voteResponses = append(voteResponses, VoteResponse{
				ID:        v.ID,
				VoterID:   v.VoterID,
				VoterName: voterName,
				VoteStage: string(v.VoteStage),
				VoteType:  string(v.VoteType),
				IsVeto:    v.IsVeto,
				Reason:    reasonPtr,
				CreatedAt: v.CreatedAt,
			})
		}
	}

	return &VoteSummaryResponse{
		Stage:        string(stage),
		TotalVoters:  totalVoters,
		Approved:     approved,
		Abstained:    abstained,
		Rejected:     rejected,
		IsTerminated: isTerminated,
		Votes:        voteResponses,
	}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// resolveVoterName loads a voter's display name from the users table.
// Returns the voter ID string if the lookup fails (non-fatal).
func (s *Service) resolveVoterName(ctx context.Context, voterID string) string {
	var u voteUserRow
	if err := s.db.WithContext(ctx).First(&u, "id = ?", voterID).Error; err != nil {
		return voterID
	}
	return u.FullName
}

// isTerminalStatus mirrors the applications package terminal-status set
// without importing it (avoids circular dependency).
func isTerminalStatus(status string) bool {
	switch status {
	case "referans_red", "yk_red", "itibar_red", "danışma_red", "yik_red", "reddedildi", "kabul":
		return true
	}
	return false
}

func (s *Service) writeLog(ctx context.Context, action, entityID, entityType string, meta map[string]interface{}) error {
	return voteWriteLogTx(ctx, s.db, action, entityID, entityType, meta)
}

func voteWriteLogTx(ctx context.Context, db *gorm.DB, action, entityID, entityType string, meta map[string]interface{}) error {
	m, _ := json.Marshal(meta)
	entry := struct {
		ID         string         `gorm:"column:id"`
		Action     string         `gorm:"column:action"`
		ActorID    *string        `gorm:"column:actor_id"`
		ActorRole  string         `gorm:"column:actor_role"`
		EntityType string         `gorm:"column:entity_type"`
		EntityID   string         `gorm:"column:entity_id"`
		Metadata   datatypes.JSON `gorm:"column:metadata"`
		CreatedAt  time.Time      `gorm:"column:created_at"`
	}{
		ID:         uuid.New().String(),
		Action:     action,
		ActorRole:  "system",
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   datatypes.JSON(m),
		CreatedAt:  time.Now(),
	}
	return db.WithContext(ctx).Table("logs").Create(&entry).Error
}
