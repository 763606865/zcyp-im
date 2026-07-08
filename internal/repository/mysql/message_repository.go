package mysql

import (
	"context"
	"database/sql"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(params repository.CreateMessageParams) (model.Message, error) {
	const query = `
INSERT INTO messages (message_no, app_id, conversation_id, sender_user_id, message_type, client_msg_id, content)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	if _, err := r.db.ExecContext(
		context.Background(),
		query,
		params.MessageNo,
		params.AppID,
		params.ConversationID,
		params.SenderUserID,
		params.MessageType,
		params.ClientMsgID,
		string(params.Content),
	); err != nil {
		return model.Message{}, err
	}

	return r.getByNo(params.MessageNo)
}

func (r *MessageRepository) ListByConversationID(conversationID uint64, limit int) ([]model.Message, error) {
	const query = `
SELECT id, message_no, app_id, conversation_id, sender_user_id, message_type, client_msg_id, content, created_at
FROM messages
WHERE conversation_id = ?
ORDER BY id DESC
LIMIT ?`

	rows, err := r.db.QueryContext(context.Background(), query, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Message, 0, limit)
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, message)
	}

	reverseMessages(items)
	return items, rows.Err()
}

func (r *MessageRepository) getByNo(messageNo string) (model.Message, error) {
	const query = `
SELECT id, message_no, app_id, conversation_id, sender_user_id, message_type, client_msg_id, content, created_at
FROM messages
WHERE message_no = ?
LIMIT 1`

	row := r.db.QueryRowContext(context.Background(), query, messageNo)
	return scanMessage(row)
}

type messageScanner interface {
	Scan(dest ...any) error
}

func scanMessage(s messageScanner) (model.Message, error) {
	var message model.Message
	var content []byte
	err := s.Scan(
		&message.ID,
		&message.MessageNo,
		&message.AppID,
		&message.ConversationID,
		&message.SenderUserID,
		&message.MessageType,
		&message.ClientMsgID,
		&content,
		&message.CreatedAt,
	)
	if err == nil {
		message.Content = append(message.Content[:0], content...)
	}
	return message, err
}

func reverseMessages(items []model.Message) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
