package model

import "time"

type ConversationMember struct {
	ID             uint64     `json:"id"`
	AppID          uint64     `json:"app_id"`
	ConversationID uint64     `json:"conversation_id"`
	MemberUserID   string     `json:"member_user_id"`
	Role           string     `json:"role"`
	Status         string     `json:"status"`
	MutedUntil     *time.Time `json:"muted_until,omitempty"`
	MicStatus      string     `json:"mic_status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
