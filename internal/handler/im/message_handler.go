package im

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/model"
	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type MessageHandler struct {
	imService   *service.IMService
	broadcaster broadcaster
}

type broadcaster interface {
	BroadcastMessage(conversationNo string, message model.Message)
}

func NewMessageHandler(imService *service.IMService, broadcaster broadcaster) *MessageHandler {
	return &MessageHandler{imService: imService, broadcaster: broadcaster}
}

func (h *MessageHandler) CreateConversation(c *gin.Context) {
	var req service.CreateConversationInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	req.AppCode = claims.AppCode
	req.OwnerUserID = claims.UserID

	conversation, err := h.imService.CreateConversation(req)
	if err != nil {
		h.writeIMError(c, err)
		return
	}

	response.Created(c, conversation)
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req service.SendMessageInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	req.AppCode = claims.AppCode
	req.SenderUserID = claims.UserID
	req.ConversationNo = c.Param("conversation_no")

	message, err := h.imService.SendMessage(req)
	if err != nil {
		h.writeIMError(c, err)
		return
	}

	if h.broadcaster != nil {
		h.broadcaster.BroadcastMessage(req.ConversationNo, message)
	}

	response.Created(c, message)
}

func (h *MessageHandler) ListMessages(c *gin.Context) {
	limit := 50
	if value := c.DefaultQuery("limit", "50"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || parsed > 200 {
			response.Error(c, http.StatusBadRequest, "limit must be between 1 and 200")
			return
		}
		limit = parsed
	}

	items, err := h.imService.ListMessages(
		mustClaims(c).AppCode,
		mustClaims(c).UserID,
		c.Param("conversation_no"),
		limit,
	)
	if err != nil {
		h.writeIMError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *MessageHandler) writeIMError(c *gin.Context, err error) {
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
	}

	response.Error(c, status, err.Error())
}
