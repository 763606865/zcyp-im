ALTER TABLE conversations
DROP INDEX uk_app_conversation_key,
DROP COLUMN conversation_key;
