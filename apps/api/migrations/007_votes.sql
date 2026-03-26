-- 007_votes.sql
-- Create votes table for YK/YIK voting with veto support
-- IMMUTABLE: No UPDATE or DELETE allowed after creation

CREATE TABLE IF NOT EXISTS votes (
    id CHAR(36) PRIMARY KEY,
    application_id CHAR(36) NOT NULL,
    voter_id CHAR(36) NOT NULL,
    
    -- Vote stage: preliminary YK, YIK, or final YK
    vote_stage ENUM('yk_prelim', 'yik', 'yk_final') NOT NULL,
    
    -- Vote type: approve, abstain, reject
    vote_type ENUM('approve', 'abstain', 'reject') NOT NULL,
    
    -- Veto flag (only valid for reject votes in specific contexts)
    is_veto BOOLEAN NOT NULL DEFAULT FALSE,
    
    reason TEXT,
    
    -- Immutable record - only created_at, no updated_at
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    FOREIGN KEY (voter_id) REFERENCES users(id) ON DELETE RESTRICT,
    
    -- One vote per voter per stage per application
    UNIQUE KEY unique_vote_per_stage (application_id, voter_id, vote_stage),
    
    INDEX idx_application_id (application_id),
    INDEX idx_voter_id (voter_id),
    INDEX idx_vote_stage (vote_stage),
    INDEX idx_vote_type (vote_type),
    INDEX idx_is_veto (is_veto)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='This table is append-only. No UPDATE or DELETE operations allowed after creation.';
