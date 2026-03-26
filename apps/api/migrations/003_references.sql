-- 003_references.sql
-- Create references table for Asil/Akademik membership verification

CREATE TABLE IF NOT EXISTS `references` (
    id CHAR(36) PRIMARY KEY,
    application_id CHAR(36) NOT NULL,
    referee_name VARCHAR(255) NOT NULL,
    referee_email VARCHAR(255) NOT NULL,
    referee_phone VARCHAR(50),
    referee_organization VARCHAR(255),
    
    -- Token security: Store SHA-256 hash only, never the raw token
    token_hash VARCHAR(64) NOT NULL,
    
    token_expires_at TIMESTAMP NOT NULL,
    token_used_at TIMESTAMP NULL,
    
    -- Reference tracking
    is_replacement BOOLEAN NOT NULL DEFAULT FALSE,
    round INT NOT NULL DEFAULT 1,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    
    INDEX idx_application_id (application_id),
    INDEX idx_token_hash (token_hash),
    INDEX idx_token_expires_at (token_expires_at),
    INDEX idx_referee_email (referee_email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
