ALTER TABLE conversation_members
ADD COLUMN muted_until TIMESTAMP NULL DEFAULT NULL AFTER status;
