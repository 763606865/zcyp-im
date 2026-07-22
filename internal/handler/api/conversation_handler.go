package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type ConversationHandler struct {
	imService *service.IMService
}

type createConversationRequest struct {
	ConversationKey string   `json:"conversation_key"`
	Type            string   `json:"type" binding:"required"`
	Scene           string   `json:"scene"`
	Subject         string   `json:"subject"`
	OwnerUserID     string   `json:"owner_user_id" binding:"required"`
	MemberUserIDs   []string `json:"member_user_ids"`
	Metadata        struct {
		ConversationKey string  `json:"conversation_key"`
		IdentityIDs     []int64 `json:"identity_ids"`
	} `json:"metadata"`
}

func NewConversationHandler(imService *service.IMService) *ConversationHandler {
	return &ConversationHandler{imService: imService}
}

func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req createConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	conversationKey := strings.TrimSpace(req.ConversationKey)
	if conversationKey == "" {
		conversationKey = strings.TrimSpace(req.Metadata.ConversationKey)
	}

	app := mustApp(c)
	conversation, err := h.imService.CreateConversation(service.CreateConversationInput{
		AppCode:         app.AppCode,
		ConversationKey: conversationKey,
		Type:            req.Type,
		Scene:           req.Scene,
		Subject:         req.Subject,
		OwnerUserID:     req.OwnerUserID,
		MemberUserIDs:   req.MemberUserIDs,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.Created(c, conversation)
}

func (h *ConversationHandler) ListMessages(c *gin.Context) {
	limit := 50
	if value := c.DefaultQuery("limit", "50"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 || parsed > 200 {
			response.Error(c, http.StatusBadRequest, "limit must be between 1 and 200")
			return
		}
		limit = parsed
	}

	userID := c.Query("user_id")
	if userID == "" {
		response.Error(c, http.StatusBadRequest, "user_id is required")
		return
	}

	app := mustApp(c)
	items, err := h.imService.ListMessages(app.AppCode, userID, c.Param("conversation_no"), limit)
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *ConversationHandler) ListMembers(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		response.Error(c, http.StatusBadRequest, "user_id is required")
		return
	}

	app := mustApp(c)
	items, err := h.imService.ListConversationMembers(app.AppCode, c.Param("conversation_no"), userID)
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}
