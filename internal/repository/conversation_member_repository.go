package repository

import "time"
import "zcyp-im/internal/model"

type CreateConversationMemberParams struct {
	AppID          uint64
	ConversationID uint64
	MemberUserID   string
	Role           string
	Status         string
}

type UpdateConversationMemberStatusParams struct {
	ConversationID uint64
	MemberUserID   string
	Status         string
}

type UpdateConversationMemberRoleParams struct {
	ConversationID uint64
	MemberUserID   string
	Role           string
}

type UpdateConversationMemberMuteParams struct {
	ConversationID uint64
	MemberUserID   string
	MutedUntil     *time.Time
}

type UpdateConversationMemberMicParams struct {
	ConversationID uint64
	MemberUserID   string
	MicStatus      string
}

type ConversationMemberRepository interface {
	Add(params CreateConversationMemberParams) error
	IsMember(conversationID uint64, memberUserID string) (bool, error)
	Get(conversationID uint64, memberUserID string) (model.ConversationMember, error)
	List(conversationID uint64) ([]model.ConversationMember, error)
	Remove(conversationID uint64, memberUserID string) error
	UpdateStatus(params UpdateConversationMemberStatusParams) error
	UpdateRole(params UpdateConversationMemberRoleParams) error
	UpdateMute(params UpdateConversationMemberMuteParams) error
	UpdateMic(params UpdateConversationMemberMicParams) error
}
