package repository

import (
	"encoding/json"

	"zcyp-im/internal/model"
)

type CreateMessageParams struct {
	MessageNo      string
	AppID          uint64
	ConversationID uint64
	SenderUserID   string
	MessageType    string
	ClientMsgID    string
	Content        json.RawMessage
}

type MessageRepository interface {
	Create(params CreateMessageParams) (model.Message, error)
	ListByConversationID(conversationID uint64, limit int) ([]model.Message, error)
}
