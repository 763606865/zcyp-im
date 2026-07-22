package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) UpsertUser(c *gin.Context) {
	var input service.UpsertUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.AppCode = c.Param("app_code")
	user, err := h.userService.UpsertUser(input)
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	user, err := h.userService.GetUser(c.Param("app_code"), c.Param("external_user_id"))
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	limit := 50
	if value := c.DefaultQuery("limit", "50"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || parsed > 200 {
			response.Error(c, http.StatusBadRequest, "limit must be between 1 and 200")
			return
		}
		limit = parsed
	}

	items, err := h.userService.ListUsers(c.Param("app_code"), limit)
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	var input service.UpdateUserStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	input.AppCode = c.Param("app_code")
	input.ExternalUserID = c.Param("external_user_id")

	user, err := h.userService.UpdateUserStatus(input)
	if err != nil {
		h.writeUserError(c, err)
		return
	}

	response.OK(c, user)
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
	}

	response.Error(c, status, err.Error())
}
