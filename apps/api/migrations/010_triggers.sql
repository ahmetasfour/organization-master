-- 010_triggers.sql
-- Create trigger to enforce immutability of rejection_reason once set

DROP TRIGGER IF EXISTS prevent_rejection_reason_update;

CREATE TRIGGER prevent_rejection_reason_update
BEFORE UPDATE ON applications
FOR EACH ROW
BEGIN
    -- If rejection_reason was previously set and is being changed, prevent the update
    IF OLD.rejection_reason IS NOT NULL 
       AND NEW.rejection_reason IS NOT NULL 
       AND OLD.rejection_reason != NEW.rejection_reason THEN
        SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'rejection_reason is immutable once set';
    END IF;
END;
