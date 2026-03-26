-- 009_logs.sql
-- Create logs table for comprehensive audit trail
-- APPEND-ONLY: No UPDATE or DELETE operations allowed

CREATE TABLE IF NOT EXISTS logs (
    id CHAR(36) PRIMARY KEY,
    
    -- Action identifier (e.g., "application.created", "vote.cast", "auth.login")
    action VARCHAR(255) NOT NULL,
    
    -- Actor information
    actor_id CHAR(36),
    actor_role VARCHAR(50),
    actor_email VARCHAR(255),
    
    -- Target entity
    entity_type VARCHAR(50),
    entity_id CHAR(36),
    
    -- Request metadata
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Additional context as JSON
    metadata JSON,
    
    -- Immutable timestamp
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_action (action),
    INDEX idx_actor_id (actor_id),
    INDEX idx_entity_type_id (entity_type, entity_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='This table is append-only. No UPDATE or DELETE operations allowed.';
