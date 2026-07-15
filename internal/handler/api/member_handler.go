package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type MemberHandler struct {
	imService *service.IMService
}

type addMembersRequest struct {
	OperatorUserID string   `json:"operator_user_id" binding:"required"`
	MemberUserIDs  []string `json:"member_user_ids" binding:"required"`
}

type updateMemberRoleRequest struct {
	OperatorUserID string `json:"operator_user_id" binding:"required"`
	Role           string `json:"role" binding:"required"`
}

type updateMemberMicRequest struct {
	OperatorUserID string `json:"operator_user_id" binding:"required"`
	MicStatus      string `json:"mic_status" binding:"required"`
}

type moderateMemberRequest struct {
	OperatorUserID string `json:"operator_user_id" binding:"required"`
}

type muteMemberRequest struct {
	OperatorUserID string `json:"operator_user_id" binding:"required"`
	Minutes        int    `json:"minutes" binding:"required"`
}

type updateConversationControlsRequest struct {
	OperatorUserID string `json:"operator_user_id" binding:"required"`
	Enabled        bool   `json:"enabled"`
}

func NewMemberHandler(imService *service.IMService) *MemberHandler {
	return &MemberHandler{imService: imService}
}

func (h *MemberHandler) AddMembers(c *gin.Context) {
	var req addMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	items, err := h.imService.AddConversationMembers(service.AddConversationMembersInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		MemberUserIDs:  req.MemberUserIDs,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *MemberHandler) RemoveMember(c *gin.Context) {
	var req moderateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	err := h.imService.RemoveConversationMember(service.RemoveConversationMemberInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		MemberUserID:   c.Param("member_user_id"),
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) BanMember(c *gin.Context) {
	h.moderateMember(c, h.imService.BanConversationMember)
}

func (h *MemberHandler) UnbanMember(c *gin.Context) {
	h.moderateMember(c, h.imService.UnbanConversationMember)
}

func (h *MemberHandler) UpdateMemberRole(c *gin.Context) {
	var req updateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	err := h.imService.UpdateConversationMemberRole(service.UpdateConversationMemberRoleInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		MemberUserID:   c.Param("member_user_id"),
		Role:           req.Role,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UpdateMemberMic(c *gin.Context) {
	var req updateMemberMicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	err := h.imService.UpdateConversationMemberMic(service.UpdateConversationMemberMicInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		MemberUserID:   c.Param("member_user_id"),
		MicStatus:      req.MicStatus,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) MuteMember(c *gin.Context) {
	var req muteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	err := h.imService.MuteConversationMember(service.MuteConversationMemberInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		MemberUserID:   c.Param("member_user_id"),
		Minutes:        req.Minutes,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UnmuteMember(c *gin.Context) {
	h.moderateMember(c, h.imService.UnmuteConversationMember)
}

func (h *MemberHandler) UpdateConversationAllMuted(c *gin.Context) {
	h.updateConversationControls(c, h.imService.UpdateConversationAllMuted)
}

func (h *MemberHandler) UpdateConversationReview(c *gin.Context) {
	h.updateConversationControls(c, h.imService.UpdateConversationReview)
}

func (h *MemberHandler) moderateMember(c *gin.Context, action func(service.ModerateConversationMemberInput) error) {
	var req moderateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	err := action(service.ModerateConversationMemberInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		MemberUserID:   c.Param("member_user_id"),
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) updateConversationControls(c *gin.Context, action func(service.UpdateConversationControlsInput) error) {
	var req updateConversationControlsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	app := mustApp(c)
	err := action(service.UpdateConversationControlsInput{
		AppCode:        app.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: req.OperatorUserID,
		Enabled:        req.Enabled,
	})
	if err != nil {
		writeIMError(c, err)
		return
	}

	response.OK(c, nil)
}
