-- 005_consultations.sql
-- Create consultations table for Profesyonel/Öğrenci membership process

CREATE TABLE IF NOT EXISTS consultations (
    id               CHAR(36)    NOT NULL PRIMARY KEY,
    application_id   CHAR(36)    NOT NULL,

    -- Member info (denormalised to avoid JOINs in public token endpoints)
    member_user_id   CHAR(36)    NOT NULL,
    member_name      VARCHAR(255) NOT NULL,
    member_email     VARCHAR(255) NOT NULL,

    -- Token security: store SHA-256 hash only, NEVER the raw token
    token_hash       VARCHAR(64)  NOT NULL,
    token_expires_at TIMESTAMP    NOT NULL,
    token_used_at    TIMESTAMP    NULL,

    -- Response (populated when member submits the form)
    response_type    ENUM('positive','negative') NULL,
    reason           TEXT,

    created_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE  KEY uq_token_hash   (token_hash),
    INDEX   idx_application_id  (application_id),
    INDEX   idx_member_user_id  (member_user_id),
    INDEX   idx_token_expires   (token_expires_at),
    INDEX   idx_response_type   (response_type),

    CONSTRAINT fk_consult_application FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE,
    CONSTRAINT fk_consult_member      FOREIGN KEY (member_user_id)  REFERENCES users(id)        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

