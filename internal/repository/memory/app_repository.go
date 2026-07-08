package memory

import (
	"fmt"
	"sync"
	"time"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

type AppRepository struct {
	mu     sync.RWMutex
	nextID uint64
	apps   map[string]model.App
}

func NewAppRepository() *AppRepository {
	return &AppRepository{
		nextID: 1,
		apps:   make(map[string]model.App),
	}
}

func (r *AppRepository) List() ([]model.App, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]model.App, 0, len(r.apps))
	for _, app := range r.apps {
		items = append(items, app)
	}

	return items, nil
}

func (r *AppRepository) Create(params repository.CreateAppParams) (model.App, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	app := model.App{
		ID:         r.nextID,
		AppCode:    params.AppCode,
		Name:       params.Name,
		AppKey:     params.AppKey,
		AppSecret:  params.AppSecret,
		Status:     params.Status,
		Scenario:   params.Scenario,
		CreatedAt:  now,
		ModifiedAt: now,
	}

	r.nextID++
	r.apps[app.AppCode] = app
	return app, nil
}

func (r *AppRepository) GetByCode(appCode string) (model.App, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	app, ok := r.apps[appCode]
	if !ok {
		return model.App{}, fmt.Errorf("app %s: %w", appCode, repository.ErrNotFound)
	}

	return app, nil
}
