package model

import "time"

type User struct {
	ID             uint64    `json:"id"`
	AppID          uint64    `json:"app_id"`
	ExternalUserID string    `json:"external_user_id"`
	Nickname       string    `json:"nickname"`
	AvatarURL      string    `json:"avatar_url"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
