-- 005_consultations.sql
-- Create consultations table for Profesyonel/Öğrenci membership process

CREATE TABLE IF NOT EXISTS consultations (
    id CHAR(36) PRIMARY KEY,
    application_id CHAR(36) NOT NULL,
    assigned_to_user_id CHAR(36) NOT NULL,
    
    notes TEXT,
    recommendation TEXT,
    is_approved BOOLEAN,
    
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_to_user_id) REFERENCES users(id) ON DELETE RESTRICT,
    
    INDEX idx_application_id (application_id),
    INDEX idx_assigned_to (assigned_to_user_id),
    INDEX idx_completed_at (completed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
