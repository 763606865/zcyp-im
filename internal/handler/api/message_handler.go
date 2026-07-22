package api

import (
	"encoding/json"
	"net/http"

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

type sendMessageRequest struct {
	SenderUserID string          `json:"sender_user_id" binding:"required"`
	MessageType  string          `json:"message_type" binding:"required"`
	ClientMsgID  string          `json:"client_msg_id"`
	Content      json.RawMessage `json:"content" binding:"required"`
}

func NewMessageHandler(imService *service.IMService, broadcaster broadcaster) *MessageHandler {
	return &MessageHandler{
		imService:   imService,
		broadcaster: broadcaster,
	}
}

func (h *MessageHandler) SendMessage(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	message, err := h.imService.SendMessage(service.SendMessageInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		SenderUserID:   req.SenderUserID,
		MessageType:    req.MessageType,
		ClientMsgID:    req.ClientMsgID,
		Content:        req.Content,
		Source:         service.SendSourceAPI,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	if h.broadcaster != nil {
		h.broadcaster.BroadcastMessage(c.Param("conversation_no"), message)
	}

	response.Created(c, message)
}
