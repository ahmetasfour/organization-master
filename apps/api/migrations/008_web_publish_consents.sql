-- 008_web_publish_consents.sql
-- Create web_publish_consents table for tracking publication permissions

CREATE TABLE IF NOT EXISTS web_publish_consents (
    id CHAR(36) PRIMARY KEY,
    application_id CHAR(36) NOT NULL,
    
    consented BOOLEAN NOT NULL,
    recorded_by CHAR(36) NULL,
    consent_given_at TIMESTAMP NULL,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    FOREIGN KEY (recorded_by) REFERENCES users(id) ON DELETE SET NULL,
    
    -- One consent record per application
    UNIQUE KEY unique_consent_per_application (application_id),
    
    INDEX idx_consented (consented)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
