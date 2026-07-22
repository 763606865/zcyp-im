package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
)

var ErrConversationNotFound = errors.New("conversation not found")
var ErrConversationAccessDenied = errors.New("conversation access denied")
var ErrConversationOwnerRequired = errors.New("conversation owner required")
var ErrConversationTypeInvalid = errors.New("conversation type invalid")
var ErrConversationMembersInvalid = errors.New("conversation members invalid")
var ErrConversationJoinNotAllowed = errors.New("conversation join not allowed")
var ErrConversationLeaveNotAllowed = errors.New("conversation leave not allowed")
var ErrConversationBanNotAllowed = errors.New("conversation ban not allowed")
var ErrConversationSpeakNotAllowed = errors.New("conversation speak not allowed")
var ErrConversationMuted = errors.New("conversation member muted")
var ErrConversationRoleInvalid = errors.New("conversation member role invalid")
var ErrConversationMicStatusInvalid = errors.New("conversation mic status invalid")
var ErrConversationReviewRejected = errors.New("conversation review rejected")
var ErrSystemConversationInvalid = errors.New("system conversation invalid")

const (
	SendSourceAPI       = "api"
	SendSourceClient    = "client"
	SendSourceWebSocket = "websocket"
)

type CreateConversationInput struct {
	AppCode         string   `json:"app_code" binding:"required"`
	ConversationKey string   `json:"conversation_key"`
	Type            string   `json:"type" binding:"required"`
	Scene           string   `json:"scene"`
	Subject         string   `json:"subject"`
	OwnerUserID     string   `json:"owner_user_id"`
	MemberUserIDs   []string `json:"member_user_ids"`
}

type SendMessageInput struct {
	AppCode        string          `json:"app_code" binding:"required"`
	ConversationNo string          `json:"conversation_no" binding:"required"`
	SenderUserID   string          `json:"sender_user_id" binding:"required"`
	MessageType    string          `json:"message_type" binding:"required"`
	ClientMsgID    string          `json:"client_msg_id"`
	Content        json.RawMessage `json:"content" binding:"required"`
	Source         string          `json:"-"`
}

type AddConversationMembersInput struct {
	AppCode        string   `json:"app_code"`
	ConversationNo string   `json:"conversation_no"`
	OperatorUserID string   `json:"operator_user_id"`
	MemberUserIDs  []string `json:"member_user_ids" binding:"required"`
}

type RemoveConversationMemberInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	OperatorUserID string `json:"operator_user_id"`
	MemberUserID   string `json:"member_user_id" binding:"required"`
}

type JoinConversationInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	UserID         string `json:"user_id"`
}

type LeaveConversationInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	UserID         string `json:"user_id"`
}

type ModerateConversationMemberInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	OperatorUserID string `json:"operator_user_id"`
	MemberUserID   string `json:"member_user_id"`
}

type UpdateConversationControlsInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	OperatorUserID string `json:"operator_user_id"`
	Enabled        bool   `json:"enabled"`
}

type UpdateConversationMemberRoleInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	OperatorUserID string `json:"operator_user_id"`
	MemberUserID   string `json:"member_user_id"`
	Role           string `json:"role"`
}

type MuteConversationMemberInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	OperatorUserID string `json:"operator_user_id"`
	MemberUserID   string `json:"member_user_id"`
	Minutes        int    `json:"minutes"`
}

type UpdateConversationMemberMicInput struct {
	AppCode        string `json:"app_code"`
	ConversationNo string `json:"conversation_no"`
	OperatorUserID string `json:"operator_user_id"`
	MemberUserID   string `json:"member_user_id"`
	MicStatus      string `json:"mic_status"`
}

type AuthenticatedUser struct {
	AppCode string
	UserID  string
}
type IMService struct {
	appService       *AppService
	userService      *UserService
	conversationRepo repository.ConversationRepository
	memberRepo       repository.ConversationMemberRepository
	messageRepo      repository.MessageRepository
	blockedWords     []string
}

func NewIMService(
	appService *AppService,
	userService *UserService,
	conversationRepo repository.ConversationRepository,
	memberRepo repository.ConversationMemberRepository,
	messageRepo repository.MessageRepository,
	blockedWords []string,
) *IMService {
	return &IMService{
		appService:       appService,
		userService:      userService,
		conversationRepo: conversationRepo,
		memberRepo:       memberRepo,
		messageRepo:      messageRepo,
		blockedWords:     blockedWords,
	}
}

func (s *IMService) CreateConversation(input CreateConversationInput) (model.Conversation, error) {
	app, err := s.appService.GetApp(input.AppCode)
	if err != nil {
		return model.Conversation{}, err
	}

	input.ConversationKey = strings.TrimSpace(input.ConversationKey)
	input.Type = strings.ToLower(strings.TrimSpace(input.Type))
	input.Scene = strings.ToLower(strings.TrimSpace(input.Scene))
	if input.ConversationKey != "" {
		conversation, err := s.conversationRepo.GetByKey(app.ID, input.ConversationKey)
		if err == nil {
			if input.Scene == "system" && (conversation.Type != "single" || conversation.Scene != "system" || conversation.OwnerUserID != input.OwnerUserID) {
				return model.Conversation{}, ErrSystemConversationInvalid
			}
			return conversation, nil
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return model.Conversation{}, err
		}
	}

	if err := validateConversationType(input.Type, input.MemberUserIDs); err != nil {
		return model.Conversation{}, err
	}

	owner, err := s.userService.GetActiveUser(input.AppCode, input.OwnerUserID)
	if err != nil {
		return model.Conversation{}, err
	}
	if err := validateConversationScene(input.Type, input.Scene, input.OwnerUserID, input.MemberUserIDs, owner); err != nil {
		return model.Conversation{}, err
	}

	memberIDs := uniqueMembers(input.OwnerUserID, input.MemberUserIDs)
	for _, memberUserID := range memberIDs {
		memberUser, err := s.userService.GetActiveUser(input.AppCode, memberUserID)
		if err != nil {
			return model.Conversation{}, err
		}
		if strings.EqualFold(input.Scene, "system") && memberUserID != input.OwnerUserID && memberUser.UserType != "normal" {
			return model.Conversation{}, ErrSystemConversationInvalid
		}
	}

	conversationNo, err := randomCode("conv", 8)
	if err != nil {
		return model.Conversation{}, err
	}

	conversation, err := s.conversationRepo.Create(repository.CreateConversationParams{
		ConversationNo:  conversationNo,
		ConversationKey: input.ConversationKey,
		AppID:           app.ID,
		Type:            input.Type,
		Scene:           input.Scene,
		Subject:         input.Subject,
		OwnerUserID:     input.OwnerUserID,
	})
	if err != nil {
		return model.Conversation{}, err
	}

	for _, memberUserID := range memberIDs {
		if err := s.memberRepo.Add(repository.CreateConversationMemberParams{
			AppID:          app.ID,
			ConversationID: conversation.ID,
			MemberUserID:   memberUserID,
			Role:           memberRole(input.Type, input.OwnerUserID, memberUserID),
			Status:         "active",
		}); err != nil {
			return model.Conversation{}, err
		}
	}

	return conversation, nil
}

func (s *IMService) SendMessage(input SendMessageInput) (model.Message, error) {
	app, conversation, err := s.resolveConversation(input.AppCode, input.ConversationNo)
	if err != nil {
		return model.Message{}, err
	}

	isMember, err := s.memberRepo.IsMember(conversation.ID, input.SenderUserID)
	if err != nil {
		return model.Message{}, err
	}
	if !isMember {
		return model.Message{}, ErrConversationAccessDenied
	}

	sender, err := s.userService.GetActiveUser(input.AppCode, input.SenderUserID)
	if err != nil {
		return model.Message{}, err
	}
	if err := validateSystemMessage(conversation, sender, input.Source, input.MessageType); err != nil {
		return model.Message{}, err
	}

	member, err := s.memberRepo.Get(conversation.ID, input.SenderUserID)
	if err != nil {
		return model.Message{}, err
	}
	if err := ensureCanSpeak(conversation.Type, conversation.AllMuted, member); err != nil {
		return model.Message{}, err
	}
	if conversation.RequireReview {
		if err := s.reviewContent(input.Content); err != nil {
			return model.Message{}, err
		}
	}
	if err := ensureMessageTypeAllowed(conversation.Type, input.MessageType); err != nil {
		return model.Message{}, err
	}

	messageNo, err := randomCode("msg", 8)
	if err != nil {
		return model.Message{}, err
	}

	content, err := buildMessageContent(input.MessageType, input.Content)
	if err != nil {
		return model.Message{}, err
	}

	return s.messageRepo.Create(repository.CreateMessageParams{
		MessageNo:      messageNo,
		AppID:          app.ID,
		ConversationID: conversation.ID,
		SenderUserID:   input.SenderUserID,
		MessageType:    input.MessageType,
		ClientMsgID:    input.ClientMsgID,
		Content:        content,
	})
}

func (s *IMService) ListMessages(appCode, userID, conversationNo string, limit int) ([]model.Message, error) {
	_, conversation, err := s.resolveConversation(appCode, conversationNo)
	if err != nil {
		return nil, err
	}

	isMember, err := s.memberRepo.IsMember(conversation.ID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrConversationAccessDenied
	}

	return s.messageRepo.ListByConversationID(conversation.ID, limit)
}

func (s *IMService) GetConversation(appCode, conversationNo string) (model.Conversation, error) {
	_, conversation, err := s.resolveConversation(appCode, conversationNo)
	if err != nil {
		return model.Conversation{}, err
	}

	return conversation, nil
}

func (s *IMService) CheckMembership(appCode, conversationNo, userID string) (model.Conversation, error) {
	_, conversation, err := s.resolveConversation(appCode, conversationNo)
	if err != nil {
		return model.Conversation{}, err
	}

	isMember, err := s.memberRepo.IsMember(conversation.ID, userID)
	if err != nil {
		return model.Conversation{}, err
	}
	if !isMember {
		return model.Conversation{}, ErrConversationAccessDenied
	}

	return conversation, nil
}

func (s *IMService) ListConversationMembers(appCode, conversationNo, userID string) ([]model.ConversationMember, error) {
	conversation, err := s.CheckMembership(appCode, conversationNo, userID)
	if err != nil {
		return nil, err
	}

	return s.memberRepo.List(conversation.ID)
}

// ListActiveConversationMemberUserIDs returns the delivery audience for an
// already persisted message. It is intended for the trusted WebSocket gateway.
func (s *IMService) ListActiveConversationMemberUserIDs(conversationID uint64) ([]string, error) {
	members, err := s.memberRepo.List(conversationID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]string, 0, len(members))
	for _, member := range members {
		if member.Status == "active" {
			userIDs = append(userIDs, member.MemberUserID)
		}
	}
	return userIDs, nil
}

func (s *IMService) AddConversationMembers(input AddConversationMembersInput) ([]model.ConversationMember, error) {
	conversation, err := s.requireOwner(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return nil, err
	}

	memberIDs := uniqueMembers("", input.MemberUserIDs)
	for _, memberUserID := range memberIDs {
		if memberUserID == "" {
			continue
		}
		if _, err := s.userService.GetActiveUser(input.AppCode, memberUserID); err != nil {
			return nil, err
		}
		if err := s.memberRepo.Add(repository.CreateConversationMemberParams{
			AppID:          conversation.AppID,
			ConversationID: conversation.ID,
			MemberUserID:   memberUserID,
			Role:           "member",
			Status:         "active",
		}); err != nil {
			return nil, err
		}
	}

	return s.memberRepo.List(conversation.ID)
}

func (s *IMService) RemoveConversationMember(input RemoveConversationMemberInput) error {
	conversation, err := s.requireOwner(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}

	if conversation.OwnerUserID == input.MemberUserID {
		return ErrConversationOwnerRequired
	}

	return s.memberRepo.Remove(conversation.ID, input.MemberUserID)
}

func (s *IMService) JoinConversation(input JoinConversationInput) error {
	conversation, err := s.GetConversation(input.AppCode, input.ConversationNo)
	if err != nil {
		return err
	}
	if !supportsSelfJoin(conversation.Type) {
		return ErrConversationJoinNotAllowed
	}
	if _, err := s.userService.GetActiveUser(input.AppCode, input.UserID); err != nil {
		return err
	}

	member, err := s.memberRepo.Get(conversation.ID, input.UserID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	if err == nil && member.Status == "banned" {
		return ErrConversationBanNotAllowed
	}

	return s.memberRepo.Add(repository.CreateConversationMemberParams{
		AppID:          conversation.AppID,
		ConversationID: conversation.ID,
		MemberUserID:   input.UserID,
		Role:           memberRole(conversation.Type, conversation.OwnerUserID, input.UserID),
		Status:         "active",
	})
}

func (s *IMService) LeaveConversation(input LeaveConversationInput) error {
	conversation, err := s.GetConversation(input.AppCode, input.ConversationNo)
	if err != nil {
		return err
	}
	if !supportsSelfJoin(conversation.Type) {
		return ErrConversationLeaveNotAllowed
	}
	if conversation.OwnerUserID == input.UserID {
		return ErrConversationOwnerRequired
	}

	return s.memberRepo.UpdateStatus(repository.UpdateConversationMemberStatusParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.UserID,
		Status:         "left",
	})
}

func (s *IMService) BanConversationMember(input ModerateConversationMemberInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	if conversation.OwnerUserID == input.MemberUserID {
		return ErrConversationOwnerRequired
	}

	return s.memberRepo.UpdateStatus(repository.UpdateConversationMemberStatusParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.MemberUserID,
		Status:         "banned",
	})
}

func (s *IMService) UnbanConversationMember(input ModerateConversationMemberInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}

	return s.memberRepo.UpdateStatus(repository.UpdateConversationMemberStatusParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.MemberUserID,
		Status:         "left",
	})
}

func (s *IMService) UpdateConversationMemberRole(input UpdateConversationMemberRoleInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	if conversation.OwnerUserID == input.MemberUserID {
		return ErrConversationOwnerRequired
	}
	if err := validateMemberRole(conversation.Type, input.Role); err != nil {
		return err
	}

	return s.memberRepo.UpdateRole(repository.UpdateConversationMemberRoleParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.MemberUserID,
		Role:           input.Role,
	})
}

func (s *IMService) MuteConversationMember(input MuteConversationMemberInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	if conversation.OwnerUserID == input.MemberUserID {
		return ErrConversationOwnerRequired
	}
	if input.Minutes <= 0 {
		return ErrConversationMembersInvalid
	}

	mutedUntil := time.Now().Add(time.Duration(input.Minutes) * time.Minute)
	return s.memberRepo.UpdateMute(repository.UpdateConversationMemberMuteParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.MemberUserID,
		MutedUntil:     &mutedUntil,
	})
}

func (s *IMService) UnmuteConversationMember(input ModerateConversationMemberInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	return s.memberRepo.UpdateMute(repository.UpdateConversationMemberMuteParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.MemberUserID,
		MutedUntil:     nil,
	})
}

func (s *IMService) UpdateConversationAllMuted(input UpdateConversationControlsInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	return s.conversationRepo.UpdateControls(repository.UpdateConversationControlsParams{
		ConversationID: conversation.ID,
		AllMuted:       &input.Enabled,
	})
}

func (s *IMService) UpdateConversationReview(input UpdateConversationControlsInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	return s.conversationRepo.UpdateControls(repository.UpdateConversationControlsParams{
		ConversationID: conversation.ID,
		RequireReview:  &input.Enabled,
	})
}

func (s *IMService) UpdateConversationMemberMic(input UpdateConversationMemberMicInput) error {
	conversation, err := s.requireModerator(input.AppCode, input.ConversationNo, input.OperatorUserID)
	if err != nil {
		return err
	}
	if strings.ToLower(conversation.Type) != "live_room" {
		return ErrConversationMicStatusInvalid
	}
	if input.MicStatus != "on" && input.MicStatus != "off" {
		return ErrConversationMicStatusInvalid
	}

	member, err := s.memberRepo.Get(conversation.ID, input.MemberUserID)
	if err != nil {
		return err
	}
	if input.MicStatus == "on" && member.Role != "owner" && member.Role != "admin" && member.Role != "speaker" {
		return ErrConversationMicStatusInvalid
	}

	return s.memberRepo.UpdateMic(repository.UpdateConversationMemberMicParams{
		ConversationID: conversation.ID,
		MemberUserID:   input.MemberUserID,
		MicStatus:      input.MicStatus,
	})
}

func (s *IMService) resolveConversation(appCode, conversationNo string) (model.App, model.Conversation, error) {
	app, err := s.appService.GetApp(appCode)
	if err != nil {
		return model.App{}, model.Conversation{}, err
	}

	conversation, err := s.conversationRepo.GetByNo(conversationNo)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.App{}, model.Conversation{}, ErrConversationNotFound
		}
		return model.App{}, model.Conversation{}, err
	}

	if conversation.AppID != app.ID {
		return model.App{}, model.Conversation{}, ErrConversationNotFound
	}

	return app, conversation, nil
}

func buildMessageContent(messageType string, content json.RawMessage) (json.RawMessage, error) {
	var payload map[string]any
	if err := json.Unmarshal(content, &payload); err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, errors.New("content must be a json object")
	}

	if strings.ToLower(messageType) == "text" {
		text, ok := payload["text"].(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil, errors.New("text message content.text is required")
		}
	}

	payload["type"] = messageType

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(raw), nil
}

func uniqueMembers(ownerUserID string, memberUserIDs []string) []string {
	seen := make(map[string]struct{})
	items := make([]string, 0, len(memberUserIDs)+1)
	if ownerUserID != "" {
		seen[ownerUserID] = struct{}{}
		items = append(items, ownerUserID)
	}
	for _, memberUserID := range memberUserIDs {
		if memberUserID == "" {
			continue
		}
		if _, ok := seen[memberUserID]; ok {
			continue
		}
		seen[memberUserID] = struct{}{}
		items = append(items, memberUserID)
	}
	return items
}

func memberRole(conversationType, ownerUserID, memberUserID string) string {
	if ownerUserID == memberUserID {
		return "owner"
	}
	if strings.ToLower(conversationType) == "live_room" {
		return "audience"
	}
	return "member"
}

func (s *IMService) requireOwner(appCode, conversationNo, userID string) (model.Conversation, error) {
	conversation, err := s.CheckMembership(appCode, conversationNo, userID)
	if err != nil {
		return model.Conversation{}, err
	}

	member, err := s.memberRepo.Get(conversation.ID, userID)
	if err != nil {
		return model.Conversation{}, err
	}
	if member.Role != "owner" {
		return model.Conversation{}, ErrConversationOwnerRequired
	}

	return conversation, nil
}

func (s *IMService) requireModerator(appCode, conversationNo, userID string) (model.Conversation, error) {
	conversation, err := s.CheckMembership(appCode, conversationNo, userID)
	if err != nil {
		return model.Conversation{}, err
	}

	member, err := s.memberRepo.Get(conversation.ID, userID)
	if err != nil {
		return model.Conversation{}, err
	}
	if member.Role != "owner" && member.Role != "admin" {
		return model.Conversation{}, ErrConversationAccessDenied
	}

	return conversation, nil
}

func validateConversationType(conversationType string, memberUserIDs []string) error {
	switch strings.ToLower(conversationType) {
	case "single":
		if len(uniqueMembers("", memberUserIDs)) != 1 {
			return ErrConversationMembersInvalid
		}
	case "group":
		return nil
	case "chatroom", "live_room":
		if len(uniqueMembers("", memberUserIDs)) != 0 {
			return ErrConversationMembersInvalid
		}
		return nil
	default:
		return ErrConversationTypeInvalid
	}

	return nil
}

func validateConversationScene(conversationType, scene, ownerUserID string, memberUserIDs []string, owner model.User) error {
	scene = strings.ToLower(strings.TrimSpace(scene))
	if scene != "system" {
		return nil
	}
	if strings.ToLower(conversationType) != "single" || owner.UserType != "system" || len(memberUserIDs) != 1 || memberUserIDs[0] == ownerUserID {
		return ErrSystemConversationInvalid
	}
	return nil
}

func validateSystemMessage(conversation model.Conversation, sender model.User, source, messageType string) error {
	if strings.ToLower(conversation.Scene) != "system" {
		return nil
	}
	if source != SendSourceAPI || sender.UserType != "system" || strings.ToLower(messageType) != "system_notice" {
		return ErrConversationSpeakNotAllowed
	}
	return nil
}

func supportsSelfJoin(conversationType string) bool {
	switch strings.ToLower(conversationType) {
	case "chatroom", "live_room":
		return true
	default:
		return false
	}
}

func ensureCanSpeak(conversationType string, allMuted bool, member model.ConversationMember) error {
	if member.Status != "active" {
		return ErrConversationAccessDenied
	}
	if member.MutedUntil != nil && member.MutedUntil.After(time.Now()) {
		return ErrConversationMuted
	}
	if allMuted && member.Role != "owner" && member.Role != "admin" {
		return ErrConversationSpeakNotAllowed
	}

	switch strings.ToLower(conversationType) {
	case "live_room":
		if member.Role != "owner" && member.Role != "admin" && member.Role != "speaker" {
			return ErrConversationSpeakNotAllowed
		}
		if member.Role == "speaker" && member.MicStatus != "on" {
			return ErrConversationSpeakNotAllowed
		}
	}

	return nil
}

func validateMemberRole(conversationType, role string) error {
	switch strings.ToLower(conversationType) {
	case "live_room":
		switch role {
		case "admin", "speaker", "audience":
			return nil
		}
	case "chatroom", "group":
		switch role {
		case "admin", "member":
			return nil
		}
	case "single":
		switch role {
		case "member":
			return nil
		}
	}
	return ErrConversationRoleInvalid
}

func (s *IMService) reviewContent(content json.RawMessage) error {
	lower := strings.ToLower(string(content))
	for _, blockedWord := range s.blockedWords {
		word := strings.TrimSpace(strings.ToLower(blockedWord))
		if word == "" {
			continue
		}
		if strings.Contains(lower, word) {
			return ErrConversationReviewRejected
		}
	}
	return nil
}

func ensureMessageTypeAllowed(conversationType, messageType string) error {
	switch strings.ToLower(conversationType) {
	case "live_room":
		if strings.ToLower(messageType) != "text" {
			return ErrConversationSpeakNotAllowed
		}
	}
	return nil
}
