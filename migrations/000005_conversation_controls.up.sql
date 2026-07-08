ALTER TABLE conversations
ADD COLUMN all_muted TINYINT(1) NOT NULL DEFAULT 0 AFTER owner_user_id,
ADD COLUMN require_review TINYINT(1) NOT NULL DEFAULT 0 AFTER all_muted;

ALTER TABLE conversation_members
ADD COLUMN mic_status VARCHAR(32) NOT NULL DEFAULT 'off' AFTER muted_until;
