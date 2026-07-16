ALTER TABLE conversations
ADD COLUMN conversation_key VARCHAR(255) NULL COMMENT '业务侧传入的会话幂等键' AFTER conversation_no,
ADD UNIQUE KEY uk_app_conversation_key (app_id, conversation_key);
