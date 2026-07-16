package model

import "time"

type Conversation struct {
	ID              uint64    `json:"id"`
	ConversationNo  string    `json:"conversation_no"`
	ConversationKey string    `json:"conversation_key"`
	AppID           uint64    `json:"app_id"`
	Type            string    `json:"type"`
	Subject         string    `json:"subject"`
	OwnerUserID     string    `json:"owner_user_id"`
	AllMuted        bool      `json:"all_muted"`
	RequireReview   bool      `json:"require_review"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
