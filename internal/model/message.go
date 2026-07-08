package model

import (
	"encoding/json"
	"time"
)

type Message struct {
	ID             uint64          `json:"id"`
	MessageNo      string          `json:"message_no"`
	AppID          uint64          `json:"app_id"`
	ConversationID uint64          `json:"conversation_id"`
	SenderUserID   string          `json:"sender_user_id"`
	MessageType    string          `json:"message_type"`
	ClientMsgID    string          `json:"client_msg_id"`
	Content        json.RawMessage `json:"content"`
	CreatedAt      time.Time       `json:"created_at"`
}
