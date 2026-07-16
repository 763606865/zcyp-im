package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository/memory"
	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type testBroadcaster struct {
	conversationNo string
	message        model.Message
	called         bool
}

func (b *testBroadcaster) BroadcastMessage(conversationNo string, message model.Message) {
	b.conversationNo = conversationNo
	b.message = message
	b.called = true
}

type envelope struct {
	Code    int             `json:"code"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

func TestMessageHandlerSendMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appService := service.NewAppService(memory.NewAppRepository())
	userService := service.NewUserService(appService, memory.NewUserRepository())
	imService := service.NewIMService(
		appService,
		userService,
		memory.NewConversationRepository(),
		memory.NewConversationMemberRepository(),
		memory.NewMessageRepository(),
		nil,
	)

	app, err := appService.CreateApp(service.CreateAppInput{Name: "test-app"})
	if err != nil {
		t.Fatalf("create app: %v", err)
	}

	if _, err := userService.UpsertUser(service.UpsertUserInput{
		AppCode:        app.AppCode,
		ExternalUserID: "u_1001",
		Nickname:       "sender",
	}); err != nil {
		t.Fatalf("upsert sender: %v", err)
	}

	conversation, err := imService.CreateConversation(service.CreateConversationInput{
		AppCode:     app.AppCode,
		Type:        "group",
		OwnerUserID: "u_1001",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	broadcaster := &testBroadcaster{}
	handler := NewMessageHandler(imService, broadcaster)

	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.POST("/conversations/:conversation_no/messages", handler.SendMessage)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/conversations/"+conversation.ConversationNo+"/messages",
		strings.NewReader(`{"sender_user_id":"u_1001","message_type":"text","client_msg_id":"client_1","content":{"text":"hello"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Code", app.AppCode)
	req.Header.Set("X-App-Key", app.AppKey)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var resp envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}

	var message model.Message
	if err := json.Unmarshal(resp.Data, &message); err != nil {
		t.Fatalf("decode message: %v", err)
	}

	if message.SenderUserID != "u_1001" {
		t.Fatalf("sender_user_id = %q", message.SenderUserID)
	}
	if message.MessageType != "text" {
		t.Fatalf("message_type = %q", message.MessageType)
	}
	var content map[string]string
	if err := json.Unmarshal(message.Content, &content); err != nil {
		t.Fatalf("decode content: %v", err)
	}
	if content["type"] != "text" || content["text"] != "hello" {
		t.Fatalf("content = %#v", content)
	}
	if !broadcaster.called {
		t.Fatal("expected broadcast to be called")
	}
	if broadcaster.conversationNo != conversation.ConversationNo {
		t.Fatalf("broadcast conversation_no = %q", broadcaster.conversationNo)
	}
	if broadcaster.message.MessageNo != message.MessageNo {
		t.Fatalf("broadcast message_no = %q, response message_no = %q", broadcaster.message.MessageNo, message.MessageNo)
	}
}

func TestMessageHandlerSendMessageRejectsNonMember(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appService := service.NewAppService(memory.NewAppRepository())
	userService := service.NewUserService(appService, memory.NewUserRepository())
	imService := service.NewIMService(
		appService,
		userService,
		memory.NewConversationRepository(),
		memory.NewConversationMemberRepository(),
		memory.NewMessageRepository(),
		nil,
	)

	app, err := appService.CreateApp(service.CreateAppInput{Name: "test-app"})
	if err != nil {
		t.Fatalf("create app: %v", err)
	}

	for _, userID := range []string{"u_1001", "u_1002"} {
		if _, err := userService.UpsertUser(service.UpsertUserInput{
			AppCode:        app.AppCode,
			ExternalUserID: userID,
		}); err != nil {
			t.Fatalf("upsert user %s: %v", userID, err)
		}
	}

	conversation, err := imService.CreateConversation(service.CreateConversationInput{
		AppCode:     app.AppCode,
		Type:        "group",
		OwnerUserID: "u_1001",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	handler := NewMessageHandler(imService, nil)
	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.POST("/conversations/:conversation_no/messages", handler.SendMessage)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/conversations/"+conversation.ConversationNo+"/messages",
		strings.NewReader(`{"sender_user_id":"u_1002","message_type":"text","content":{"text":"hello"}}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Code", app.AppCode)
	req.Header.Set("X-App-Key", app.AppKey)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var resp envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if resp.Message != "conversation access denied" {
		t.Fatalf("message = %q", resp.Message)
	}
}

func TestMessageHandlerSendBizCardMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	appService := service.NewAppService(memory.NewAppRepository())
	userService := service.NewUserService(appService, memory.NewUserRepository())
	imService := service.NewIMService(
		appService,
		userService,
		memory.NewConversationRepository(),
		memory.NewConversationMemberRepository(),
		memory.NewMessageRepository(),
		nil,
	)

	app, err := appService.CreateApp(service.CreateAppInput{Name: "test-app"})
	if err != nil {
		t.Fatalf("create app: %v", err)
	}

	if _, err := userService.UpsertUser(service.UpsertUserInput{
		AppCode:        app.AppCode,
		ExternalUserID: "u_1001",
	}); err != nil {
		t.Fatalf("upsert sender: %v", err)
	}

	conversation, err := imService.CreateConversation(service.CreateConversationInput{
		AppCode:     app.AppCode,
		Type:        "group",
		OwnerUserID: "u_1001",
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	handler := NewMessageHandler(imService, nil)
	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.POST("/conversations/:conversation_no/messages", handler.SendMessage)

	body := `{"sender_user_id":"u_1001","message_type":"biz_card","client_msg_id":"client_2","content":{"card_type":"oa_approval","title":"审批申请","summary":"张三提交了请假申请","status":"pending","action_url":"/oa/approvals/12345","biz_id":"approval_12345"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/conversations/"+conversation.ConversationNo+"/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Code", app.AppCode)
	req.Header.Set("X-App-Key", app.AppKey)

	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}

	var resp envelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}

	var message model.Message
	if err := json.Unmarshal(resp.Data, &message); err != nil {
		t.Fatalf("decode message: %v", err)
	}

	var content map[string]any
	if err := json.Unmarshal(message.Content, &content); err != nil {
		t.Fatalf("decode content: %v", err)
	}
	if content["type"] != "biz_card" {
		t.Fatalf("type = %#v", content["type"])
	}
	if content["card_type"] != "oa_approval" {
		t.Fatalf("card_type = %#v", content["card_type"])
	}
	if content["biz_id"] != "approval_12345" {
		t.Fatalf("biz_id = %#v", content["biz_id"])
	}
}
