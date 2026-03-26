-- 002_applications.sql
-- Create applications table with complete state machine status values

CREATE TABLE IF NOT EXISTS applications (
    id CHAR(36) PRIMARY KEY,
    applicant_name VARCHAR(255) NOT NULL,
    applicant_email VARCHAR(255) NOT NULL,
    applicant_phone VARCHAR(50),
    applicant_bio TEXT,
    membership_type ENUM('asil', 'akademik', 'profesyonel', 'öğrenci', 'onursal') NOT NULL,
    
    -- Status with ALL possible values from state machine
    status ENUM(
        'başvuru_alındı',
        'referans_bekleniyor',
        'referans_tamamlandı',
        'referans_red',
        'yk_ön_incelemede',
        'ön_onaylandı',
        'yk_red',
        'itibar_taramasında',
        'itibar_temiz',
        'itibar_red',
        'danışma_sürecinde',
        'danışma_red',
        'öneri_alındı',
        'yik_değerlendirmede',
        'yik_red',
        'gündemde',
        'kabul',
        'reddedildi'
    ) NOT NULL DEFAULT 'başvuru_alındı',
    
    -- Onursal-specific fields
    proposed_by_user_id CHAR(36),
    proposal_reason TEXT,
    
    -- Rejection tracking (WRITE-ONCE field)
    rejection_reason TEXT,
    rejected_by_role VARCHAR(50),
    
    -- Web publishing
    web_publish_consent BOOLEAN NOT NULL DEFAULT FALSE,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Re-application tracking
    previous_app_id CHAR(36),
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (proposed_by_user_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (previous_app_id) REFERENCES applications(id) ON DELETE SET NULL,
    
    INDEX idx_status (status),
    INDEX idx_membership_type (membership_type),
    INDEX idx_applicant_email (applicant_email),
    INDEX idx_proposed_by (proposed_by_user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
