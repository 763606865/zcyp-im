ALTER TABLE conversation_members
ADD COLUMN muted_until TIMESTAMP NULL DEFAULT NULL COMMENT '禁言截止时间' AFTER status;
