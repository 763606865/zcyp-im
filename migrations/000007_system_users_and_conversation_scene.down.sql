ALTER TABLE conversations
    DROP KEY idx_app_scene,
    DROP COLUMN scene;

ALTER TABLE im_users
    DROP KEY idx_app_user_type,
    DROP COLUMN user_type;
