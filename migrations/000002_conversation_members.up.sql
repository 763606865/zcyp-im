CREATE TABLE IF NOT EXISTS conversation_members (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    app_id BIGINT UNSIGNED NOT NULL,
    conversation_id BIGINT UNSIGNED NOT NULL,
    member_user_id VARCHAR(64) NOT NULL,
    role VARCHAR(32) NOT NULL DEFAULT 'member',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_conv_member (conversation_id, member_user_id),
    KEY idx_app_member (app_id, member_user_id)
);
