package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type ConversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

func (r *ConversationRepository) Create(params repository.CreateConversationParams) (model.Conversation, error) {
	const query = `
INSERT INTO conversations (conversation_no, conversation_key, app_id, type, scene, subject, owner_user_id)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	if _, err := r.db.ExecContext(
		context.Background(),
		query,
		params.ConversationNo,
		nullableString(params.ConversationKey),
		params.AppID,
		params.Type,
		params.Scene,
		params.Subject,
		params.OwnerUserID,
	); err != nil {
		return model.Conversation{}, err
	}

	return r.GetByNo(params.ConversationNo)
}

func (r *ConversationRepository) GetByNo(conversationNo string) (model.Conversation, error) {
	const query = `
SELECT id, conversation_no, conversation_key, app_id, type, scene, subject, owner_user_id, all_muted, require_review, created_at, updated_at
FROM conversations
WHERE conversation_no = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, conversationNo)
	return scanConversation(row, fmt.Sprintf("conversation %s", conversationNo))
}

func (r *ConversationRepository) GetByKey(appID uint64, conversationKey string) (model.Conversation, error) {
	const query = `
SELECT id, conversation_no, conversation_key, app_id, type, scene, subject, owner_user_id, all_muted, require_review, created_at, updated_at
FROM conversations
WHERE app_id = ? AND conversation_key = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, appID, conversationKey)
	return scanConversation(row, fmt.Sprintf("conversation key %s", conversationKey))
}

func (r *ConversationRepository) UpdateControls(params repository.UpdateConversationControlsParams) error {
	query := `
UPDATE conversations
SET all_muted = CASE WHEN ? IS NULL THEN all_muted ELSE ? END,
    require_review = CASE WHEN ? IS NULL THEN require_review ELSE ? END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?`

	var allMutedValue any
	var requireReviewValue any
	if params.AllMuted != nil {
		allMutedValue = *params.AllMuted
	}
	if params.RequireReview != nil {
		requireReviewValue = *params.RequireReview
	}

	_, err := r.db.ExecContext(context.Background(), query, allMutedValue, allMutedValue, requireReviewValue, requireReviewValue, params.ConversationID)
	return err
}

type conversationScanner interface {
	Scan(dest ...any) error
}

func scanConversation(s conversationScanner, entity string) (model.Conversation, error) {
	var conversation model.Conversation
	var conversationKey sql.NullString
	if err := s.Scan(
		&conversation.ID,
		&conversation.ConversationNo,
		&conversationKey,
		&conversation.AppID,
		&conversation.Type,
		&conversation.Scene,
		&conversation.Subject,
		&conversation.OwnerUserID,
		&conversation.AllMuted,
		&conversation.RequireReview,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.Conversation{}, fmt.Errorf("%s: %w", entity, repository.ErrNotFound)
		}
		return model.Conversation{}, err
	}
	if conversationKey.Valid {
		conversation.ConversationKey = conversationKey.String
	}
	return conversation, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
