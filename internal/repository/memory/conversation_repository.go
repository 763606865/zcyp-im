package memory

import (
	"fmt"
	"sync"
	"time"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type ConversationRepository struct {
	mu            sync.RWMutex
	nextID        uint64
	conversations map[string]model.Conversation
}

func NewConversationRepository() *ConversationRepository {
	return &ConversationRepository{
		nextID:        1,
		conversations: make(map[string]model.Conversation),
	}
}

func (r *ConversationRepository) Create(params repository.CreateConversationParams) (model.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	conversation := model.Conversation{
		ID:             r.nextID,
		ConversationNo: params.ConversationNo,
		AppID:          params.AppID,
		Type:           params.Type,
		Subject:        params.Subject,
		OwnerUserID:    params.OwnerUserID,
		AllMuted:       false,
		RequireReview:  false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	r.nextID++
	r.conversations[conversation.ConversationNo] = conversation

	return conversation, nil
}

func (r *ConversationRepository) UpdateControls(params repository.UpdateConversationControlsParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, conversation := range r.conversations {
		if conversation.ID != params.ConversationID {
			continue
		}
		if params.AllMuted != nil {
			conversation.AllMuted = *params.AllMuted
		}
		if params.RequireReview != nil {
			conversation.RequireReview = *params.RequireReview
		}
		conversation.UpdatedAt = time.Now()
		r.conversations[key] = conversation
		return nil
	}

	return repository.ErrNotFound
}

func (r *ConversationRepository) GetByNo(conversationNo string) (model.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversation, ok := r.conversations[conversationNo]
	if !ok {
		return model.Conversation{}, fmt.Errorf("conversation %s: %w", conversationNo, repository.ErrNotFound)
	}

	return conversation, nil
}
