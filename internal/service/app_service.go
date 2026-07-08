package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

var ErrAppNotFound = errors.New("app not found")

type CreateAppInput struct {
	Name     string   `json:"name" binding:"required"`
	Scenario []string `json:"scenario"`
}

type AppService struct {
	repo repository.AppRepository
}

func NewAppService(repo repository.AppRepository) *AppService {
	return &AppService{
		repo: repo,
	}
}

func (s *AppService) ListApps() ([]model.App, error) {
	return s.repo.List()
}

func (s *AppService) CreateApp(input CreateAppInput) (model.App, error) {
	appCode, err := randomCode("app", 8)
	if err != nil {
		return model.App{}, err
	}

	appKey, err := randomHex(16)
	if err != nil {
		return model.App{}, err
	}

	appSecret, err := randomHex(24)
	if err != nil {
		return model.App{}, err
	}

	return s.repo.Create(repository.CreateAppParams{
		AppCode:   appCode,
		Name:      input.Name,
		AppKey:    appKey,
		AppSecret: appSecret,
		Status:    "active",
		Scenario:  input.Scenario,
	})
}

func (s *AppService) GetApp(appCode string) (model.App, error) {
	app, err := s.repo.GetByCode(appCode)
	if err != nil {
		return model.App{}, ErrAppNotFound
	}

	return app, nil
}

func (s *AppService) ValidateApp(appCode, appKey string) (model.App, error) {
	app, err := s.GetApp(appCode)
	if err != nil {
		return model.App{}, err
	}

	if app.AppKey != appKey {
		return model.App{}, ErrAppNotFound
	}

	return app, nil
}

func randomHex(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func randomCode(prefix string, size int) (string, error) {
	suffix, err := randomHex(size)
	if err != nil {
		return "", err
	}

	return prefix + "_" + suffix, nil
}
