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

func TestConversationHandlerCreateConversation(t *testing.T) {
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

	handler := NewConversationHandler(imService)
	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.POST("/conversations", handler.CreateConversation)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/conversations",
		strings.NewReader(`{"type":"group","subject":"项目群","owner_user_id":"u_1001","member_user_ids":["u_1002"]}`),
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

	var conversation model.Conversation
	if err := json.Unmarshal(resp.Data, &conversation); err != nil {
		t.Fatalf("decode conversation: %v", err)
	}

	if conversation.Type != "group" {
		t.Fatalf("type = %q", conversation.Type)
	}
	if conversation.OwnerUserID != "u_1001" {
		t.Fatalf("owner_user_id = %q", conversation.OwnerUserID)
	}
}

func TestConversationHandlerListMessagesAndMembers(t *testing.T) {
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
		AppCode:       app.AppCode,
		Type:          "group",
		OwnerUserID:   "u_1001",
		MemberUserIDs: []string{"u_1002"},
	})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	if _, err := imService.SendMessage(service.SendMessageInput{
		AppCode:        app.AppCode,
		ConversationNo: conversation.ConversationNo,
		SenderUserID:   "u_1001",
		MessageType:    "text",
		Content:        "hello",
	}); err != nil {
		t.Fatalf("send message: %v", err)
	}

	handler := NewConversationHandler(imService)
	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.GET("/conversations/:conversation_no/messages", handler.ListMessages)
	apiGroup.GET("/conversations/:conversation_no/members", handler.ListMembers)

	messageReq := httptest.NewRequest(
		http.MethodGet,
		"/api/conversations/"+conversation.ConversationNo+"/messages?user_id=u_1002&limit=10",
		nil,
	)
	messageReq.Header.Set("X-App-Code", app.AppCode)
	messageReq.Header.Set("X-App-Key", app.AppKey)

	messageRecorder := httptest.NewRecorder()
	engine.ServeHTTP(messageRecorder, messageReq)

	if messageRecorder.Code != http.StatusOK {
		t.Fatalf("message status = %d, body = %s", messageRecorder.Code, messageRecorder.Body.String())
	}

	var messageResp struct {
		Code int `json:"code"`
		Data struct {
			Items []model.Message `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(messageRecorder.Body.Bytes(), &messageResp); err != nil {
		t.Fatalf("decode message response: %v", err)
	}
	if len(messageResp.Data.Items) != 1 {
		t.Fatalf("message items = %d", len(messageResp.Data.Items))
	}

	memberReq := httptest.NewRequest(
		http.MethodGet,
		"/api/conversations/"+conversation.ConversationNo+"/members?user_id=u_1001",
		nil,
	)
	memberReq.Header.Set("X-App-Code", app.AppCode)
	memberReq.Header.Set("X-App-Key", app.AppKey)

	memberRecorder := httptest.NewRecorder()
	engine.ServeHTTP(memberRecorder, memberReq)

	if memberRecorder.Code != http.StatusOK {
		t.Fatalf("member status = %d, body = %s", memberRecorder.Code, memberRecorder.Body.String())
	}

	var memberResp struct {
		Code int `json:"code"`
		Data struct {
			Items []model.ConversationMember `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(memberRecorder.Body.Bytes(), &memberResp); err != nil {
		t.Fatalf("decode member response: %v", err)
	}
	if len(memberResp.Data.Items) != 2 {
		t.Fatalf("member items = %d", len(memberResp.Data.Items))
	}
}
