ALTER TABLE conversations
MODIFY COLUMN owner_user_id VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会话创建者用户标识';

ALTER TABLE messages
MODIFY COLUMN sender_user_id VARCHAR(255) NOT NULL COMMENT '消息发送者用户标识';

ALTER TABLE conversation_members
MODIFY COLUMN member_user_id VARCHAR(255) NOT NULL COMMENT '成员用户标识';
