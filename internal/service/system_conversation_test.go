package service

import (
	"encoding/json"
	"errors"
	"testing"

	"zcyp-im/internal/repository/memory"
)

func TestSystemConversationOnlyAcceptsAPIMessageFromSystemUser(t *testing.T) {
	appService := NewAppService(memory.NewAppRepository())
	userService := NewUserService(appService, memory.NewUserRepository())
	imService := NewIMService(
		appService,
		userService,
		memory.NewConversationRepository(),
		memory.NewConversationMemberRepository(),
		memory.NewMessageRepository(),
		nil,
	)

	app, err := appService.CreateApp(CreateAppInput{Name: "test-app"})
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	for _, input := range []UpsertUserInput{
		{AppCode: app.AppCode, ExternalUserID: "system", UserType: "system"},
		{AppCode: app.AppCode, ExternalUserID: "rc_user_im:123"},
	} {
		if _, err := userService.UpsertUser(input); err != nil {
			t.Fatalf("upsert user %s: %v", input.ExternalUserID, err)
		}
	}
	if _, err := userService.GetTokenEligibleUser(app.AppCode, "system"); !errors.Is(err, ErrSystemUserTokenNotAllowed) {
		t.Fatalf("system token eligibility error = %v", err)
	}

	conversation, err := imService.CreateConversation(CreateConversationInput{
		AppCode:         app.AppCode,
		ConversationKey: "system:rc_user_im:123",
		Type:            "single",
		Scene:           "system",
		OwnerUserID:     "system",
		MemberUserIDs:   []string{"rc_user_im:123"},
	})
	if err != nil {
		t.Fatalf("create system conversation: %v", err)
	}
	if conversation.Scene != "system" {
		t.Fatalf("scene = %q", conversation.Scene)
	}

	content := json.RawMessage(`{"notice_type":"approval_created","title":"新的审批申请"}`)
	message, err := imService.SendMessage(SendMessageInput{
		AppCode:        app.AppCode,
		ConversationNo: conversation.ConversationNo,
		SenderUserID:   "system",
		MessageType:    "system_notice",
		ClientMsgID:    "approval_created_12345",
		Content:        content,
		Source:         SendSourceAPI,
	})
	if err != nil {
		t.Fatalf("send system notice: %v", err)
	}
	if message.MessageType != "system_notice" {
		t.Fatalf("message_type = %q", message.MessageType)
	}
	audience, err := imService.ListActiveConversationMemberUserIDs(conversation.ID)
	if err != nil {
		t.Fatalf("list delivery audience: %v", err)
	}
	if len(audience) != 2 {
		t.Fatalf("delivery audience = %#v", audience)
	}

	tests := []struct {
		name        string
		sender      string
		messageType string
		source      string
	}{
		{name: "client source", sender: "system", messageType: "system_notice", source: SendSourceClient},
		{name: "normal sender", sender: "rc_user_im:123", messageType: "system_notice", source: SendSourceAPI},
		{name: "wrong message type", sender: "system", messageType: "text", source: SendSourceAPI},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := imService.SendMessage(SendMessageInput{
				AppCode:        app.AppCode,
				ConversationNo: conversation.ConversationNo,
				SenderUserID:   tt.sender,
				MessageType:    tt.messageType,
				Content:        content,
				Source:         tt.source,
			})
			if !errors.Is(err, ErrConversationSpeakNotAllowed) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestSystemConversationRequiresSystemOwner(t *testing.T) {
	appService := NewAppService(memory.NewAppRepository())
	userService := NewUserService(appService, memory.NewUserRepository())
	imService := NewIMService(
		appService,
		userService,
		memory.NewConversationRepository(),
		memory.NewConversationMemberRepository(),
		memory.NewMessageRepository(),
		nil,
	)
	app, err := appService.CreateApp(CreateAppInput{Name: "test-app"})
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	for _, userID := range []string{"u_1", "u_2"} {
		if _, err := userService.UpsertUser(UpsertUserInput{AppCode: app.AppCode, ExternalUserID: userID}); err != nil {
			t.Fatalf("upsert user: %v", err)
		}
	}

	_, err = imService.CreateConversation(CreateConversationInput{
		AppCode:       app.AppCode,
		Type:          "single",
		Scene:         "system",
		OwnerUserID:   "u_1",
		MemberUserIDs: []string{"u_2"},
	})
	if !errors.Is(err, ErrSystemConversationInvalid) {
		t.Fatalf("error = %v", err)
	}
}
