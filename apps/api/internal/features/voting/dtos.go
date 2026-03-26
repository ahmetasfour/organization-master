package voting

import "time"

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// CastVoteRequest holds the payload for casting a vote.
type CastVoteRequest struct {
	// VoteType must be one of: approve, abstain, reject
	VoteType string `json:"vote_type" validate:"required,oneof=approve abstain reject"`

	// Reason is mandatory when VoteType is "reject" (minimum 20 characters).
	// For other vote types it is silently ignored.
	Reason string `json:"reason" validate:"omitempty,min=20"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// VoteResponse represents a single cast vote returned to clients.
// The Reason field is only populated for yk / admin requestors.
type VoteResponse struct {
	ID        string    `json:"id"`
	VoterID   string    `json:"voter_id"`
	VoterName string    `json:"voter_name"`
	VoteStage string    `json:"vote_stage"`
	VoteType  string    `json:"vote_type"`
	IsVeto    bool      `json:"is_veto"`
	Reason    *string   `json:"reason,omitempty"` // nil when hidden from caller
	CreatedAt time.Time `json:"created_at"`
}

// VoteSummaryResponse aggregates votes for a given stage on an application.
type VoteSummaryResponse struct {
	Stage        string         `json:"stage"`
	TotalVoters  int64          `json:"total_voters"`
	Approved     int64          `json:"approved"`
	Abstained    int64          `json:"abstained"`
	Rejected     int64          `json:"rejected"`
	IsTerminated bool           `json:"is_terminated"`
	Votes        []VoteResponse `json:"votes"` // populated for yk / admin only
}
