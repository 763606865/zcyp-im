ALTER TABLE im_users
    ADD COLUMN user_type VARCHAR(32) NOT NULL DEFAULT 'normal' COMMENT '用户类型：normal普通用户，system系统用户' AFTER avatar_url,
    ADD KEY idx_app_user_type (app_id, user_type);

ALTER TABLE conversations
    ADD COLUMN scene VARCHAR(32) NOT NULL DEFAULT '' COMMENT '会话业务场景，例如system系统通知' AFTER type,
    ADD KEY idx_app_scene (app_id, scene);
