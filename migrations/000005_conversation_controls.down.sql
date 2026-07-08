ALTER TABLE conversation_members
DROP COLUMN mic_status;

ALTER TABLE conversations
DROP COLUMN require_review,
DROP COLUMN all_muted;
