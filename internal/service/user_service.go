package service

import (
	"errors"
	"strings"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserDisabled = errors.New("user disabled")
var ErrUserStatusInvalid = errors.New("user status invalid")
var ErrUserTypeInvalid = errors.New("user type invalid")
var ErrSystemUserTokenNotAllowed = errors.New("system user token not allowed")

type UpsertUserInput struct {
	AppCode        string `json:"app_code" binding:"required"`
	ExternalUserID string `json:"external_user_id" binding:"required"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	UserType       string `json:"user_type"`
}

type UpdateUserStatusInput struct {
	AppCode        string `json:"app_code"`
	ExternalUserID string `json:"external_user_id"`
	Status         string `json:"status" binding:"required"`
}

type UpdateUserProfileInput struct {
	AppCode        string  `json:"app_code"`
	ExternalUserID string  `json:"external_user_id"`
	Nickname       *string `json:"nickname"`
	AvatarURL      *string `json:"avatar_url"`
}

type UserService struct {
	appService *AppService
	repo       repository.UserRepository
}

func NewUserService(appService *AppService, repo repository.UserRepository) *UserService {
	return &UserService{
		appService: appService,
		repo:       repo,
	}
}

func (s *UserService) UpsertUser(input UpsertUserInput) (model.User, error) {
	app, err := s.appService.GetApp(input.AppCode)
	if err != nil {
		return model.User{}, err
	}

	status := "active"
	userType := strings.ToLower(strings.TrimSpace(input.UserType))
	userTypeProvided := userType != ""
	if userType == "" {
		userType = "normal"
	}
	if userType != "normal" && userType != "system" {
		return model.User{}, ErrUserTypeInvalid
	}
	existing, err := s.repo.GetByExternalUserID(app.ID, input.ExternalUserID)
	if err == nil {
		status = existing.Status
		if !userTypeProvided {
			userType = existing.UserType
		}
	}

	return s.repo.Upsert(repository.UpsertUserParams{
		AppID:          app.ID,
		ExternalUserID: input.ExternalUserID,
		Nickname:       input.Nickname,
		AvatarURL:      input.AvatarURL,
		UserType:       userType,
		Status:         status,
	})
}

func (s *UserService) GetUser(appCode, externalUserID string) (model.User, error) {
	app, err := s.appService.GetApp(appCode)
	if err != nil {
		return model.User{}, err
	}

	user, err := s.repo.GetByExternalUserID(app.ID, externalUserID)
	if err != nil {
		return model.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetActiveUser(appCode, externalUserID string) (model.User, error) {
	user, err := s.GetUser(appCode, externalUserID)
	if err != nil {
		return model.User{}, err
	}
	if user.Status != "active" {
		return model.User{}, ErrUserDisabled
	}
	return user, nil
}

func (s *UserService) GetTokenEligibleUser(appCode, externalUserID string) (model.User, error) {
	user, err := s.GetActiveUser(appCode, externalUserID)
	if err != nil {
		return model.User{}, err
	}
	if user.UserType == "system" {
		return model.User{}, ErrSystemUserTokenNotAllowed
	}
	return user, nil
}

func (s *UserService) ListUsers(appCode string, limit int) ([]model.User, error) {
	app, err := s.appService.GetApp(appCode)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByAppID(app.ID, limit)
}

func (s *UserService) UpdateUserStatus(input UpdateUserStatusInput) (model.User, error) {
	if input.Status != "active" && input.Status != "disabled" {
		return model.User{}, ErrUserStatusInvalid
	}

	user, err := s.GetUser(input.AppCode, input.ExternalUserID)
	if err != nil {
		return model.User{}, err
	}

	return s.repo.Upsert(repository.UpsertUserParams{
		AppID:          user.AppID,
		ExternalUserID: user.ExternalUserID,
		Nickname:       user.Nickname,
		AvatarURL:      user.AvatarURL,
		UserType:       user.UserType,
		Status:         input.Status,
	})
}

func (s *UserService) UpdateUserProfile(input UpdateUserProfileInput) (model.User, error) {
	user, err := s.GetUser(input.AppCode, input.ExternalUserID)
	if err != nil {
		return model.User{}, err
	}

	nickname := user.Nickname
	if input.Nickname != nil {
		nickname = *input.Nickname
	}

	avatarURL := user.AvatarURL
	if input.AvatarURL != nil {
		avatarURL = *input.AvatarURL
	}

	return s.repo.Upsert(repository.UpsertUserParams{
		AppID:          user.AppID,
		ExternalUserID: user.ExternalUserID,
		Nickname:       nickname,
		AvatarURL:      avatarURL,
		UserType:       user.UserType,
		Status:         user.Status,
	})
}
