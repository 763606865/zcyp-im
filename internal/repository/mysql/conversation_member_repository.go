package mysql

import (
	"context"
	"database/sql"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type ConversationMemberRepository struct {
	db *sql.DB
}

func NewConversationMemberRepository(db *sql.DB) *ConversationMemberRepository {
	return &ConversationMemberRepository{db: db}
}

func (r *ConversationMemberRepository) Add(params repository.CreateConversationMemberParams) error {
	const query = `
INSERT INTO conversation_members (app_id, conversation_id, member_user_id, role, status)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE role = VALUES(role), status = VALUES(status), updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.ExecContext(
		context.Background(),
		query,
		params.AppID,
		params.ConversationID,
		params.MemberUserID,
		params.Role,
		params.Status,
	)
	return err
}

func (r *ConversationMemberRepository) IsMember(conversationID uint64, memberUserID string) (bool, error) {
	const query = `
SELECT 1
FROM conversation_members
WHERE conversation_id = ? AND member_user_id = ? AND status = 'active'
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, conversationID, memberUserID)
	var marker int
	err := row.Scan(&marker)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ConversationMemberRepository) Get(conversationID uint64, memberUserID string) (model.ConversationMember, error) {
	const query = `
SELECT id, app_id, conversation_id, member_user_id, role, status, muted_until, mic_status, created_at, updated_at
FROM conversation_members
WHERE conversation_id = ? AND member_user_id = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, conversationID, memberUserID)
	var member model.ConversationMember
	err := row.Scan(
		&member.ID,
		&member.AppID,
		&member.ConversationID,
		&member.MemberUserID,
		&member.Role,
		&member.Status,
		&member.MutedUntil,
		&member.MicStatus,
		&member.CreatedAt,
		&member.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return model.ConversationMember{}, repository.ErrNotFound
	}
	return member, err
}

func (r *ConversationMemberRepository) List(conversationID uint64) ([]model.ConversationMember, error) {
	const query = `
SELECT id, app_id, conversation_id, member_user_id, role, status, muted_until, mic_status, created_at, updated_at
FROM conversation_members
WHERE conversation_id = ?
ORDER BY id ASC`

	rows, err := r.db.QueryContext(context.Background(), query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ConversationMember, 0)
	for rows.Next() {
		var member model.ConversationMember
		if err := rows.Scan(
			&member.ID,
			&member.AppID,
			&member.ConversationID,
			&member.MemberUserID,
			&member.Role,
			&member.Status,
			&member.MutedUntil,
			&member.MicStatus,
			&member.CreatedAt,
			&member.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, member)
	}
	return items, rows.Err()
}

func (r *ConversationMemberRepository) Remove(conversationID uint64, memberUserID string) error {
	const query = `
DELETE FROM conversation_members
WHERE conversation_id = ? AND member_user_id = ?`

	_, err := r.db.ExecContext(context.Background(), query, conversationID, memberUserID)
	return err
}

func (r *ConversationMemberRepository) UpdateStatus(params repository.UpdateConversationMemberStatusParams) error {
	const query = `
UPDATE conversation_members
SET status = ?, updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ? AND member_user_id = ?`

	_, err := r.db.ExecContext(context.Background(), query, params.Status, params.ConversationID, params.MemberUserID)
	return err
}

func (r *ConversationMemberRepository) UpdateRole(params repository.UpdateConversationMemberRoleParams) error {
	const query = `
UPDATE conversation_members
SET role = ?, updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ? AND member_user_id = ?`

	_, err := r.db.ExecContext(context.Background(), query, params.Role, params.ConversationID, params.MemberUserID)
	return err
}

func (r *ConversationMemberRepository) UpdateMute(params repository.UpdateConversationMemberMuteParams) error {
	const query = `
UPDATE conversation_members
SET muted_until = ?, updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ? AND member_user_id = ?`

	_, err := r.db.ExecContext(context.Background(), query, params.MutedUntil, params.ConversationID, params.MemberUserID)
	return err
}

func (r *ConversationMemberRepository) UpdateMic(params repository.UpdateConversationMemberMicParams) error {
	const query = `
UPDATE conversation_members
SET mic_status = ?, updated_at = CURRENT_TIMESTAMP
WHERE conversation_id = ? AND member_user_id = ?`

	_, err := r.db.ExecContext(context.Background(), query, params.MicStatus, params.ConversationID, params.MemberUserID)
	return err
}
