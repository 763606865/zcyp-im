package repository

import "zcyp-im/internal/model"

type CreateConversationParams struct {
	ConversationNo string
	AppID          uint64
	Type           string
	Subject        string
	OwnerUserID    string
}

type UpdateConversationControlsParams struct {
	ConversationID uint64
	AllMuted       *bool
	RequireReview  *bool
}

type ConversationRepository interface {
	Create(params CreateConversationParams) (model.Conversation, error)
	GetByNo(conversationNo string) (model.Conversation, error)
	UpdateControls(params UpdateConversationControlsParams) error
}
