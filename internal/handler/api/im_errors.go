package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

func writeIMError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrAppNotFound):
		status = http.StatusUnauthorized
	case errors.Is(err, service.ErrUserNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrUserDisabled):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationAccessDenied):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrConversationTypeInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrSystemConversationInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationMembersInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationJoinNotAllowed):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationLeaveNotAllowed):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationBanNotAllowed):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationSpeakNotAllowed):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationMuted):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationRoleInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationMicStatusInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationReviewRejected):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationOwnerRequired):
		status = http.StatusForbidden
	}

	response.Error(c, status, err.Error())
}
