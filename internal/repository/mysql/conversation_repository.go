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
INSERT INTO conversations (conversation_no, app_id, type, subject, owner_user_id)
VALUES (?, ?, ?, ?, ?)`

	if _, err := r.db.ExecContext(
		context.Background(),
		query,
		params.ConversationNo,
		params.AppID,
		params.Type,
		params.Subject,
		params.OwnerUserID,
	); err != nil {
		return model.Conversation{}, err
	}

	return r.GetByNo(params.ConversationNo)
}

func (r *ConversationRepository) GetByNo(conversationNo string) (model.Conversation, error) {
	const query = `
SELECT id, conversation_no, app_id, type, subject, owner_user_id, all_muted, require_review, created_at, updated_at
FROM conversations
WHERE conversation_no = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, conversationNo)

	var conversation model.Conversation
	if err := row.Scan(
		&conversation.ID,
		&conversation.ConversationNo,
		&conversation.AppID,
		&conversation.Type,
		&conversation.Subject,
		&conversation.OwnerUserID,
		&conversation.AllMuted,
		&conversation.RequireReview,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return model.Conversation{}, fmt.Errorf("conversation %s: %w", conversationNo, repository.ErrNotFound)
		}
		return model.Conversation{}, err
	}

	return conversation, nil
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
