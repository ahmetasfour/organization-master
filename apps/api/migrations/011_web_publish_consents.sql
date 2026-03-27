-- Web Publish Consents Migration
-- Records whether an accepted member consents to being published on the public member list

CREATE TABLE IF NOT EXISTS web_publish_consents (
    id              CHAR(36) PRIMARY KEY,
    application_id  CHAR(36) NOT NULL,
    consented       BOOLEAN NOT NULL,
    recorded_by     CHAR(36) NOT NULL,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_application_consent (application_id),
    
    CONSTRAINT fk_consent_application 
        FOREIGN KEY (application_id) 
        REFERENCES applications(id) 
        ON DELETE RESTRICT,
    
    CONSTRAINT fk_consent_recorded_by 
        FOREIGN KEY (recorded_by) 
        REFERENCES users(id) 
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Index for faster lookups
CREATE INDEX idx_consent_application ON web_publish_consents(application_id);
CREATE INDEX idx_consent_recorded_by ON web_publish_consents(recorded_by);