-- 006_reputation_contacts.sql
-- Create reputation_contacts table for Asil/Akademik reputation screening

CREATE TABLE IF NOT EXISTS reputation_contacts (
    id CHAR(36) PRIMARY KEY,
    application_id CHAR(36) NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    contact_organization VARCHAR(255),
    
    -- Response type: clean or negative
    response_type ENUM('clean', 'negative') NOT NULL,
    
    notes TEXT,
    contacted_by_user_id CHAR(36),
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    FOREIGN KEY (contacted_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_application_id (application_id),
    INDEX idx_response_type (response_type),
    INDEX idx_contacted_by (contacted_by_user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
