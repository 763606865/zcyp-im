package memory

import (
	"fmt"
	"sync"
	"time"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type UserRepository struct {
	mu     sync.RWMutex
	nextID uint64
	users  map[string]model.User
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		nextID: 1,
		users:  make(map[string]model.User),
	}
}

func (r *UserRepository) Upsert(params repository.UpsertUserParams) (model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := r.key(params.AppID, params.ExternalUserID)
	now := time.Now()
	user, ok := r.users[key]
	if ok {
		user.Nickname = params.Nickname
		user.AvatarURL = params.AvatarURL
		user.Status = params.Status
		user.UpdatedAt = now
		r.users[key] = user
		return user, nil
	}

	user = model.User{
		ID:             r.nextID,
		AppID:          params.AppID,
		ExternalUserID: params.ExternalUserID,
		Nickname:       params.Nickname,
		AvatarURL:      params.AvatarURL,
		Status:         params.Status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.nextID++
	r.users[key] = user
	return user, nil
}

func (r *UserRepository) GetByExternalUserID(appID uint64, externalUserID string) (model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[r.key(appID, externalUserID)]
	if !ok {
		return model.User{}, fmt.Errorf("user %s: %w", externalUserID, repository.ErrNotFound)
	}
	return user, nil
}

func (r *UserRepository) ListByAppID(appID uint64, limit int) ([]model.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.User, 0, limit)
	for _, user := range r.users {
		if user.AppID != appID {
			continue
		}
		items = append(items, user)
		if limit > 0 && len(items) >= limit {
			break
		}
	}
	return items, nil
}

func (r *UserRepository) key(appID uint64, externalUserID string) string {
	return fmt.Sprintf("%d:%s", appID, externalUserID)
}
