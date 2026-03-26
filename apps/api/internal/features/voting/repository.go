package voting

import (
	"context"

	"gorm.io/gorm"
)

// Repository handles database operations for the voting system.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new voting repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create persists a new Vote record.
func (r *Repository) Create(ctx context.Context, vote *Vote) error {
	return r.db.WithContext(ctx).Create(vote).Error
}

// FindByApplicationAndStage retrieves all votes for a given application and stage.
func (r *Repository) FindByApplicationAndStage(
	ctx context.Context,
	applicationID string,
	stage VoteStage,
) ([]*Vote, error) {
	var votes []*Vote
	err := r.db.WithContext(ctx).
		Where("application_id = ? AND vote_stage = ?", applicationID, stage).
		Order("created_at ASC").
		Find(&votes).Error
	if err != nil {
		return nil, err
	}
	return votes, nil
}

// FindByApplicationVoterStage retrieves a single vote for the given combination.
// Returns (nil, nil) when no record exists (not an error).
func (r *Repository) FindByApplicationVoterStage(
	ctx context.Context,
	applicationID, voterID string,
	stage VoteStage,
) (*Vote, error) {
	var vote Vote
	err := r.db.WithContext(ctx).
		Where("application_id = ? AND voter_id = ? AND vote_stage = ?",
			applicationID, voterID, stage).
		First(&vote).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &vote, nil
}

// CountActiveVotersByRole returns the number of active (is_active = true) users
// who hold the given role. This is used to determine when all eligible voters
// have cast their ballots.
func (r *Repository) CountActiveVotersByRole(ctx context.Context, role string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("users").
		Where("role = ? AND is_active = ?", role, true).
		Count(&count).Error
	return count, err
}

// CountVotesByStageAndType counts votes with a specific type for an application stage.
func (r *Repository) CountVotesByStageAndType(
	ctx context.Context,
	applicationID string,
	stage VoteStage,
	voteType VoteType,
) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&Vote{}).
		Where("application_id = ? AND vote_stage = ? AND vote_type = ?",
			applicationID, stage, voteType).
		Count(&count).Error
	return count, err
}
