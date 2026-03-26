-- 004_reference_responses.sql
-- Create reference_responses table for referee submissions

CREATE TABLE IF NOT EXISTS reference_responses (
    id CHAR(36) PRIMARY KEY,
    reference_id CHAR(36) NOT NULL,
    
    -- Response type: positive, unknown, negative
    response_type ENUM('positive', 'unknown', 'negative') NOT NULL,
    
    relationship_description TEXT,
    duration_years INT,
    recommendation_text TEXT NOT NULL,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (reference_id) REFERENCES `references`(id) ON DELETE CASCADE,
    
    UNIQUE KEY unique_response_per_reference (reference_id),
    INDEX idx_response_type (response_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
