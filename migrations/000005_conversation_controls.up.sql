ALTER TABLE conversations
ADD COLUMN all_muted TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启全员禁言' AFTER owner_user_id,
ADD COLUMN require_review TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否开启消息审核' AFTER all_muted;

ALTER TABLE conversation_members
ADD COLUMN mic_status VARCHAR(32) NOT NULL DEFAULT 'off' COMMENT '成员麦克风状态' AFTER muted_until;
