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
	keyIndex      map[string]string
}

func NewConversationRepository() *ConversationRepository {
	return &ConversationRepository{
		nextID:        1,
		conversations: make(map[string]model.Conversation),
		keyIndex:      make(map[string]string),
	}
}

func (r *ConversationRepository) Create(params repository.CreateConversationParams) (model.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	conversation := model.Conversation{
		ID:              r.nextID,
		ConversationNo:  params.ConversationNo,
		ConversationKey: params.ConversationKey,
		AppID:           params.AppID,
		Type:            params.Type,
		Scene:           params.Scene,
		Subject:         params.Subject,
		OwnerUserID:     params.OwnerUserID,
		AllMuted:        false,
		RequireReview:   false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	r.nextID++
	r.conversations[conversation.ConversationNo] = conversation
	if conversation.ConversationKey != "" {
		r.keyIndex[r.key(conversation.AppID, conversation.ConversationKey)] = conversation.ConversationNo
	}

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

func (r *ConversationRepository) GetByKey(appID uint64, conversationKey string) (model.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversationNo, ok := r.keyIndex[r.key(appID, conversationKey)]
	if !ok {
		return model.Conversation{}, fmt.Errorf("conversation key %s: %w", conversationKey, repository.ErrNotFound)
	}

	conversation, ok := r.conversations[conversationNo]
	if !ok {
		return model.Conversation{}, fmt.Errorf("conversation key %s: %w", conversationKey, repository.ErrNotFound)
	}

	return conversation, nil
}

func (r *ConversationRepository) key(appID uint64, conversationKey string) string {
	return fmt.Sprintf("%d:%s", appID, conversationKey)
}
