package im

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type MemberHandler struct {
	imService *service.IMService
}

func NewMemberHandler(imService *service.IMService) *MemberHandler {
	return &MemberHandler{imService: imService}
}

func (h *MemberHandler) ListMembers(c *gin.Context) {
	claims := mustClaims(c)
	items, err := h.imService.ListConversationMembers(claims.AppCode, c.Param("conversation_no"), claims.UserID)
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *MemberHandler) AddMembers(c *gin.Context) {
	var req service.AddConversationMembersInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	req.AppCode = claims.AppCode
	req.OperatorUserID = claims.UserID
	req.ConversationNo = c.Param("conversation_no")

	items, err := h.imService.AddConversationMembers(req)
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, gin.H{"items": items})
}

func (h *MemberHandler) RemoveMember(c *gin.Context) {
	var req service.RemoveConversationMemberInput
	req.MemberUserID = c.Param("member_user_id")

	claims := mustClaims(c)
	req.AppCode = claims.AppCode
	req.OperatorUserID = claims.UserID
	req.ConversationNo = c.Param("conversation_no")

	if err := h.imService.RemoveConversationMember(req); err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) JoinConversation(c *gin.Context) {
	claims := mustClaims(c)
	err := h.imService.JoinConversation(service.JoinConversationInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		UserID:         claims.UserID,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) LeaveConversation(c *gin.Context) {
	claims := mustClaims(c)
	err := h.imService.LeaveConversation(service.LeaveConversationInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		UserID:         claims.UserID,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) BanMember(c *gin.Context) {
	claims := mustClaims(c)
	err := h.imService.BanConversationMember(service.ModerateConversationMemberInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		MemberUserID:   c.Param("member_user_id"),
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UnbanMember(c *gin.Context) {
	claims := mustClaims(c)
	err := h.imService.UnbanConversationMember(service.ModerateConversationMemberInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		MemberUserID:   c.Param("member_user_id"),
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UpdateMemberRole(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	err := h.imService.UpdateConversationMemberRole(service.UpdateConversationMemberRoleInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		MemberUserID:   c.Param("member_user_id"),
		Role:           req.Role,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) MuteMember(c *gin.Context) {
	var req struct {
		Minutes int `json:"minutes" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	err := h.imService.MuteConversationMember(service.MuteConversationMemberInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		MemberUserID:   c.Param("member_user_id"),
		Minutes:        req.Minutes,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UnmuteMember(c *gin.Context) {
	claims := mustClaims(c)
	err := h.imService.UnmuteConversationMember(service.ModerateConversationMemberInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		MemberUserID:   c.Param("member_user_id"),
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UpdateConversationAllMuted(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	err := h.imService.UpdateConversationAllMuted(service.UpdateConversationControlsInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		Enabled:        req.Enabled,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UpdateConversationReview(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	err := h.imService.UpdateConversationReview(service.UpdateConversationControlsInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		Enabled:        req.Enabled,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) UpdateMemberMic(c *gin.Context) {
	var req struct {
		MicStatus string `json:"mic_status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	claims := mustClaims(c)
	err := h.imService.UpdateConversationMemberMic(service.UpdateConversationMemberMicInput{
		AppCode:        claims.AppCode,
		ConversationNo: c.Param("conversation_no"),
		OperatorUserID: claims.UserID,
		MemberUserID:   c.Param("member_user_id"),
		MicStatus:      req.MicStatus,
	})
	if err != nil {
		h.writeMemberError(c, err)
		return
	}

	response.OK(c, nil)
}

func (h *MemberHandler) writeMemberError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, service.ErrConversationAccessDenied):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationOwnerRequired):
		status = http.StatusForbidden
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
	case errors.Is(err, service.ErrConversationNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrAppNotFound):
		status = http.StatusUnauthorized
	case errors.Is(err, service.ErrUserNotFound):
		status = http.StatusNotFound
	case errors.Is(err, service.ErrUserDisabled):
		status = http.StatusForbidden
	case errors.Is(err, service.ErrConversationTypeInvalid):
		status = http.StatusBadRequest
	case errors.Is(err, service.ErrConversationMembersInvalid):
		status = http.StatusBadRequest
	}

	response.Error(c, status, err.Error())
}
