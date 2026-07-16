ALTER TABLE conversation_members
MODIFY COLUMN member_user_id VARCHAR(64) NOT NULL;

ALTER TABLE messages
MODIFY COLUMN sender_user_id VARCHAR(64) NOT NULL;

ALTER TABLE conversations
MODIFY COLUMN owner_user_id VARCHAR(64) NOT NULL DEFAULT '';
