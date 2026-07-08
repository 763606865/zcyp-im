package memory

import (
	"fmt"
	"sync"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type ConversationMemberRepository struct {
	mu      sync.RWMutex
	nextID  uint64
	members map[string]model.ConversationMember
}

func NewConversationMemberRepository() *ConversationMemberRepository {
	return &ConversationMemberRepository{
		nextID:  1,
		members: make(map[string]model.ConversationMember),
	}
}

func (r *ConversationMemberRepository) Add(params repository.CreateConversationMemberParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(params.ConversationID, params.MemberUserID)
	member, ok := r.members[key]
	if ok {
		member.Role = params.Role
		member.Status = params.Status
		r.members[key] = member
		return nil
	}

	r.members[key] = model.ConversationMember{
		ID:             r.nextID,
		AppID:          params.AppID,
		ConversationID: params.ConversationID,
		MemberUserID:   params.MemberUserID,
		Role:           params.Role,
		Status:         params.Status,
		MicStatus:      "off",
	}
	r.nextID++
	return nil
}

func (r *ConversationMemberRepository) IsMember(conversationID uint64, memberUserID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.members[r.key(conversationID, memberUserID)]
	return ok, nil
}

func (r *ConversationMemberRepository) Get(conversationID uint64, memberUserID string) (model.ConversationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	member, ok := r.members[r.key(conversationID, memberUserID)]
	if !ok {
		return model.ConversationMember{}, fmt.Errorf("member %s: %w", memberUserID, repository.ErrNotFound)
	}
	return member, nil
}

func (r *ConversationMemberRepository) List(conversationID uint64) ([]model.ConversationMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.ConversationMember, 0)
	for _, member := range r.members {
		if member.ConversationID == conversationID {
			items = append(items, member)
		}
	}
	return items, nil
}

func (r *ConversationMemberRepository) Remove(conversationID uint64, memberUserID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.members, r.key(conversationID, memberUserID))
	return nil
}

func (r *ConversationMemberRepository) UpdateStatus(params repository.UpdateConversationMemberStatusParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(params.ConversationID, params.MemberUserID)
	member, ok := r.members[key]
	if !ok {
		return repository.ErrNotFound
	}
	member.Status = params.Status
	r.members[key] = member
	return nil
}

func (r *ConversationMemberRepository) UpdateRole(params repository.UpdateConversationMemberRoleParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(params.ConversationID, params.MemberUserID)
	member, ok := r.members[key]
	if !ok {
		return repository.ErrNotFound
	}
	member.Role = params.Role
	r.members[key] = member
	return nil
}

func (r *ConversationMemberRepository) UpdateMute(params repository.UpdateConversationMemberMuteParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(params.ConversationID, params.MemberUserID)
	member, ok := r.members[key]
	if !ok {
		return repository.ErrNotFound
	}
	member.MutedUntil = params.MutedUntil
	r.members[key] = member
	return nil
}

func (r *ConversationMemberRepository) UpdateMic(params repository.UpdateConversationMemberMicParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(params.ConversationID, params.MemberUserID)
	member, ok := r.members[key]
	if !ok {
		return repository.ErrNotFound
	}
	member.MicStatus = params.MicStatus
	r.members[key] = member
	return nil
}

func (r *ConversationMemberRepository) key(conversationID uint64, memberUserID string) string {
	return fmt.Sprintf("%d:%s", conversationID, memberUserID)
}
