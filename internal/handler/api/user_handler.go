package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/auth"
	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type UserHandler struct {
	userService  *service.UserService
	tokenService *auth.TokenService
}

type upsertUserRequest struct {
	ExternalUserID string `json:"external_user_id" binding:"required"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	UserType       string `json:"user_type"`
}

type updateUserRequest struct {
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
}

func NewUserHandler(userService *service.UserService, tokenService *auth.TokenService) *UserHandler {
	return &UserHandler{
		userService:  userService,
		tokenService: tokenService,
	}
}

func (h *UserHandler) UpsertUser(c *gin.Context) {
	var req upsertUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	user, err := h.userService.UpsertUser(service.UpsertUserInput{
		AppCode:        app.AppCode,
		ExternalUserID: req.ExternalUserID,
		Nickname:       req.Nickname,
		AvatarURL:      req.AvatarURL,
		UserType:       req.UserType,
	})
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	app := mustApp(c)
	user, err := h.userService.GetUser(app.AppCode, c.Param("external_user_id"))
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	user, err := h.userService.UpdateUserProfile(service.UpdateUserProfileInput{
		AppCode:        app.AppCode,
		ExternalUserID: c.Param("external_user_id"),
		Nickname:       req.Nickname,
		AvatarURL:      req.AvatarURL,
	})
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) IssueAccessToken(c *gin.Context) {
	app := mustApp(c)
	userID := c.Param("external_user_id")

	if _, err := h.userService.GetTokenEligibleUser(app.AppCode, userID); err != nil {
		h.writeUserError(c, err)
		return
	}

	token, expiresAt, err := h.tokenService.Issue(app.AppCode, userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, gin.H{
		"app_code":   app.AppCode,
		"user_id":    userID,
		"token":      token,
		"expires_at": expiresAt,
	})
}

func (h *UserHandler) writeUserError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrAppNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrUserNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrUserDisabled):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrUserStatusInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrUserTypeInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrSystemUserTokenNotAllowed):
		status = http.StatusForbidden
	}

	response.Error(c, status, err.Error())
}
