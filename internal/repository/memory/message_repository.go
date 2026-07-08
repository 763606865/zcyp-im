package memory

import (
	"encoding/json"
	"sync"
	"time"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type MessageRepository struct {
	mu       sync.RWMutex
	nextID   uint64
	messages map[uint64][]model.Message
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{
		nextID:   1,
		messages: make(map[uint64][]model.Message),
	}
}

func (r *MessageRepository) Create(params repository.CreateMessageParams) (model.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	message := model.Message{
		ID:             r.nextID,
		MessageNo:      params.MessageNo,
		AppID:          params.AppID,
		ConversationID: params.ConversationID,
		SenderUserID:   params.SenderUserID,
		MessageType:    params.MessageType,
		ClientMsgID:    params.ClientMsgID,
		Content:        json.RawMessage(append([]byte(nil), params.Content...)),
		CreatedAt:      time.Now(),
	}

	r.nextID++
	r.messages[params.ConversationID] = append(r.messages[params.ConversationID], message)

	return message, nil
}

func (r *MessageRepository) ListByConversationID(conversationID uint64, limit int) ([]model.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := r.messages[conversationID]
	if len(items) <= limit {
		copied := make([]model.Message, len(items))
		copy(copied, items)
		return copied, nil
	}

	start := len(items) - limit
	copied := make([]model.Message, limit)
	copy(copied, items[start:])
	return copied, nil
}
