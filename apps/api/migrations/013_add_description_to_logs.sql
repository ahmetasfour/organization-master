-- 013_add_description_to_logs.sql
-- Add human-readable description field to logs table

ALTER TABLE logs ADD COLUMN description VARCHAR(500) AFTER action;
ALTER TABLE logs ADD INDEX idx_description (description);
