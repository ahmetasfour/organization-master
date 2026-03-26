package voting

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoteStage represents the voting stage in the application process
type VoteStage string

const (
	VoteStageYKPrelim VoteStage = "yk_prelim" // YK preliminary vote
	VoteStageYIK      VoteStage = "yik"       // YIK vote (for Onursal only)
	VoteStageYKFinal  VoteStage = "yk_final"  // YK final vote
)

// VoteType represents the voter's decision
type VoteType string

const (
	VoteApprove VoteType = "approve"
	VoteAbstain VoteType = "abstain"
	VoteReject  VoteType = "reject"
)

// Vote represents a single vote cast by a YK or YIK member
// This is an IMMUTABLE record - once created, it cannot be modified or deleted
type Vote struct {
	ID            string    `gorm:"type:char(36);primaryKey" json:"id"`
	ApplicationID string    `gorm:"type:char(36);not null;index;uniqueIndex:unique_vote_per_stage,priority:1" json:"application_id"`
	VoterID       string    `gorm:"type:char(36);not null;index;uniqueIndex:unique_vote_per_stage,priority:2" json:"voter_id"`
	VoteStage     VoteStage `gorm:"type:enum('yk_prelim','yik','yk_final');not null;index;uniqueIndex:unique_vote_per_stage,priority:3" json:"vote_stage"`
	VoteType      VoteType  `gorm:"type:enum('approve','abstain','reject');not null;index" json:"vote_type"`
	IsVeto        bool      `gorm:"default:false;not null;index" json:"is_veto"`
	Reason        string    `gorm:"type:text" json:"reason,omitempty"`
	CreatedAt     time.Time `gorm:"not null;autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for GORM
func (Vote) TableName() string {
	return "votes"
}

// BeforeCreate GORM hook to generate UUID before creating a vote
func (v *Vote) BeforeCreate(tx *gorm.DB) error {
	if v.ID == "" {
		v.ID = uuid.New().String()
	}
	return nil
}

// IsApproval checks if the vote is an approval
func (v *Vote) IsApproval() bool {
	return v.VoteType == VoteApprove
}

// IsRejection checks if the vote is a rejection
func (v *Vote) IsRejection() bool {
	return v.VoteType == VoteReject
}

// IsAbstention checks if the vote is an abstention
func (v *Vote) IsAbstention() bool {
	return v.VoteType == VoteAbstain
}

// IsVetoVote checks if this is a veto vote
func (v *Vote) IsVetoVote() bool {
	return v.IsVeto && v.IsRejection()
}
