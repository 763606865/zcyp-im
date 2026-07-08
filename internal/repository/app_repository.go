package repository

import "zcyp-im/internal/model"

type CreateAppParams struct {
	AppCode   string
	Name      string
	AppKey    string
	AppSecret string
	Status    string
	Scenario  []string
}

type AppRepository interface {
	List() ([]model.App, error)
	Create(params CreateAppParams) (model.App, error)
	GetByCode(appCode string) (model.App, error)
}
