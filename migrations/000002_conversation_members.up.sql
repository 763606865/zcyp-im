CREATE TABLE IF NOT EXISTS conversation_members (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '会话成员主键ID',
    app_id BIGINT UNSIGNED NOT NULL COMMENT '所属应用ID',
    conversation_id BIGINT UNSIGNED NOT NULL COMMENT '所属会话ID',
    member_user_id VARCHAR(255) NOT NULL COMMENT '成员用户标识',
    role VARCHAR(32) NOT NULL DEFAULT 'member' COMMENT '成员角色',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_conv_member (conversation_id, member_user_id),
    KEY idx_app_member (app_id, member_user_id)
);
