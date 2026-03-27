-- 012_logs_immutability.sql
-- Add triggers to enforce append-only pattern on logs table

-- Prevent UPDATE on logs table
DROP TRIGGER IF EXISTS prevent_logs_update;

CREATE TRIGGER prevent_logs_update
BEFORE UPDATE ON logs
FOR EACH ROW
BEGIN
    SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'Logs table is append-only: UPDATE operations are not allowed';
END;

-- Prevent DELETE on logs table (except for system maintenance)
-- Note: This trigger can be temporarily disabled for maintenance if needed
DROP TRIGGER IF EXISTS prevent_logs_delete;

CREATE TRIGGER prevent_logs_delete
BEFORE DELETE ON logs
FOR EACH ROW
BEGIN
    SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'Logs table is append-only: DELETE operations are not allowed';
END;
