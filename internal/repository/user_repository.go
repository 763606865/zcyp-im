package repository

import "zcyp-im/internal/model"

type UpsertUserParams struct {
	AppID          uint64
	ExternalUserID string
	Nickname       string
	AvatarURL      string
	Status         string
}

type UserRepository interface {
	Upsert(params UpsertUserParams) (model.User, error)
	GetByExternalUserID(appID uint64, externalUserID string) (model.User, error)
	ListByAppID(appID uint64, limit int) ([]model.User, error)
}
