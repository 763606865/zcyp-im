CREATE TABLE IF NOT EXISTS apps (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '应用主键ID',
    app_code VARCHAR(32) NOT NULL UNIQUE COMMENT '应用编码',
    name VARCHAR(128) NOT NULL COMMENT '应用名称',
    app_key VARCHAR(64) NOT NULL UNIQUE COMMENT '应用访问Key',
    app_secret VARCHAR(96) NOT NULL COMMENT '应用访问Secret',
    status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '应用状态',
    scenario JSON NOT NULL COMMENT '支持的会话场景配置',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
);

CREATE TABLE IF NOT EXISTS im_users (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '用户主键ID',
    app_id BIGINT UNSIGNED NOT NULL COMMENT '所属应用ID',
    external_user_id VARCHAR(255) NOT NULL COMMENT '业务侧用户唯一标识',
    nickname VARCHAR(64) NOT NULL DEFAULT '' COMMENT '用户昵称',
    avatar_url VARCHAR(255) NOT NULL DEFAULT '' COMMENT '用户头像地址',
    status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '用户状态',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_app_user (app_id, external_user_id)
);

CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '会话主键ID',
    conversation_no VARCHAR(32) NOT NULL UNIQUE COMMENT '会话唯一编号',
    conversation_key VARCHAR(255) NULL COMMENT '业务侧传入的会话幂等键',
    app_id BIGINT UNSIGNED NOT NULL COMMENT '所属应用ID',
    type VARCHAR(32) NOT NULL COMMENT '会话类型',
    subject VARCHAR(128) NOT NULL DEFAULT '' COMMENT '会话标题或主题',
    owner_user_id VARCHAR(255) NOT NULL DEFAULT '' COMMENT '会话创建者用户标识',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    UNIQUE KEY uk_app_conversation_key (app_id, conversation_key),
    KEY idx_app_type (app_id, type)
);

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT COMMENT '消息主键ID',
    message_no VARCHAR(32) NOT NULL UNIQUE COMMENT '消息唯一编号',
    app_id BIGINT UNSIGNED NOT NULL COMMENT '所属应用ID',
    conversation_id BIGINT UNSIGNED NOT NULL COMMENT '所属会话ID',
    sender_user_id VARCHAR(255) NOT NULL COMMENT '消息发送者用户标识',
    message_type VARCHAR(32) NOT NULL COMMENT '消息类型',
    client_msg_id VARCHAR(64) NOT NULL DEFAULT '' COMMENT '客户端消息唯一标识',
    content JSON NOT NULL COMMENT '消息内容JSON',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    KEY idx_conv_created (conversation_id, created_at),
    KEY idx_app_client_msg (app_id, client_msg_id)
);
