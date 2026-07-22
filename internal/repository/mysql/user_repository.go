package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Upsert(params repository.UpsertUserParams) (model.User, error) {
	const query = `
INSERT INTO im_users (app_id, external_user_id, nickname, avatar_url, user_type, status)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
nickname = VALUES(nickname),
avatar_url = VALUES(avatar_url),
user_type = VALUES(user_type),
status = VALUES(status),
updated_at = CURRENT_TIMESTAMP`

	if _, err := r.db.ExecContext(
		context.Background(),
		query,
		params.AppID,
		params.ExternalUserID,
		params.Nickname,
		params.AvatarURL,
		params.UserType,
		params.Status,
	); err != nil {
		return model.User{}, err
	}

	return r.GetByExternalUserID(params.AppID, params.ExternalUserID)
}

func (r *UserRepository) GetByExternalUserID(appID uint64, externalUserID string) (model.User, error) {
	const query = `
SELECT id, app_id, external_user_id, nickname, avatar_url, user_type, status, created_at, updated_at
FROM im_users
WHERE app_id = ? AND external_user_id = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, appID, externalUserID)
	user, err := scanUser(row)
	if err == sql.ErrNoRows {
		return model.User{}, fmt.Errorf("user %s: %w", externalUserID, repository.ErrNotFound)
	}
	return user, err
}

func (r *UserRepository) ListByAppID(appID uint64, limit int) ([]model.User, error) {
	const query = `
SELECT id, app_id, external_user_id, nickname, avatar_url, user_type, status, created_at, updated_at
FROM im_users
WHERE app_id = ?
ORDER BY id DESC
LIMIT ?`

	rows, err := r.db.QueryContext(context.Background(), query, appID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.User, 0, limit)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	return items, rows.Err()
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(s userScanner) (model.User, error) {
	var user model.User
	err := s.Scan(
		&user.ID,
		&user.AppID,
		&user.ExternalUserID,
		&user.Nickname,
		&user.AvatarURL,
		&user.UserType,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}
