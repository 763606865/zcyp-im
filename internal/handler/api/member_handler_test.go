package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/model"
	"zcyp-im/internal/repository/memory"
	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

func TestMemberHandlerAddAndRemoveMembers(t *testing.T) {
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

	for _, userID := range []string{"u_1001", "u_1002", "u_1003"} {
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

	handler := NewMemberHandler(imService)
	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.POST("/conversations/:conversation_no/members", handler.AddMembers)
	apiGroup.DELETE("/conversations/:conversation_no/members/:member_user_id", handler.RemoveMember)

	addReq := httptest.NewRequest(
		http.MethodPost,
		"/api/conversations/"+conversation.ConversationNo+"/members",
		strings.NewReader(`{"operator_user_id":"u_1001","member_user_ids":["u_1003"]}`),
	)
	addReq.Header.Set("Content-Type", "application/json")
	addReq.Header.Set("X-App-Code", app.AppCode)
	addReq.Header.Set("X-App-Key", app.AppKey)

	addRecorder := httptest.NewRecorder()
	engine.ServeHTTP(addRecorder, addReq)

	if addRecorder.Code != http.StatusOK {
		t.Fatalf("add status = %d, body = %s", addRecorder.Code, addRecorder.Body.String())
	}

	var addResp struct {
		Code int `json:"code"`
		Data struct {
			Items []model.ConversationMember `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(addRecorder.Body.Bytes(), &addResp); err != nil {
		t.Fatalf("decode add response: %v", err)
	}
	if len(addResp.Data.Items) != 3 {
		t.Fatalf("member items after add = %d", len(addResp.Data.Items))
	}

	removeReq := httptest.NewRequest(
		http.MethodDelete,
		"/api/conversations/"+conversation.ConversationNo+"/members/u_1003",
		strings.NewReader(`{"operator_user_id":"u_1001"}`),
	)
	removeReq.Header.Set("Content-Type", "application/json")
	removeReq.Header.Set("X-App-Code", app.AppCode)
	removeReq.Header.Set("X-App-Key", app.AppKey)

	removeRecorder := httptest.NewRecorder()
	engine.ServeHTTP(removeRecorder, removeReq)

	if removeRecorder.Code != http.StatusOK {
		t.Fatalf("remove status = %d, body = %s", removeRecorder.Code, removeRecorder.Body.String())
	}

	items, err := imService.ListConversationMembers(app.AppCode, conversation.ConversationNo, "u_1001")
	if err != nil {
		t.Fatalf("list members: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("member items after remove = %d", len(items))
	}
}

func TestMemberHandlerConversationControls(t *testing.T) {
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

	handler := NewMemberHandler(imService)
	engine := gin.New()
	engine.Use(response.Middleware())
	apiGroup := engine.Group("/api")
	apiGroup.Use(AppAuthMiddleware(appService))
	apiGroup.POST("/conversations/:conversation_no/members/:member_user_id/mute", handler.MuteMember)
	apiGroup.POST("/conversations/:conversation_no/members/:member_user_id/unmute", handler.UnmuteMember)
	apiGroup.POST("/conversations/:conversation_no/all-muted", handler.UpdateConversationAllMuted)
	apiGroup.POST("/conversations/:conversation_no/review", handler.UpdateConversationReview)
	apiGroup.POST("/conversations/:conversation_no/members/:member_user_id/ban", handler.BanMember)

	for _, tc := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodPost, "/api/conversations/" + conversation.ConversationNo + "/members/u_1002/mute", `{"operator_user_id":"u_1001","minutes":10}`},
		{http.MethodPost, "/api/conversations/" + conversation.ConversationNo + "/all-muted", `{"operator_user_id":"u_1001","enabled":true}`},
		{http.MethodPost, "/api/conversations/" + conversation.ConversationNo + "/review", `{"operator_user_id":"u_1001","enabled":true}`},
		{http.MethodPost, "/api/conversations/" + conversation.ConversationNo + "/members/u_1002/ban", `{"operator_user_id":"u_1001"}`},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-App-Code", app.AppCode)
		req.Header.Set("X-App-Key", app.AppKey)

		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("request %s %s status = %d, body = %s", tc.method, tc.path, recorder.Code, recorder.Body.String())
		}
	}

	updatedConversation, err := imService.GetConversation(app.AppCode, conversation.ConversationNo)
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if !updatedConversation.AllMuted {
		t.Fatal("expected all_muted to be enabled")
	}
	if !updatedConversation.RequireReview {
		t.Fatal("expected require_review to be enabled")
	}

	members, err := imService.ListConversationMembers(app.AppCode, conversation.ConversationNo, "u_1001")
	if err != nil {
		t.Fatalf("list members: %v", err)
	}

	var target model.ConversationMember
	found := false
	for _, item := range members {
		if item.MemberUserID == "u_1002" {
			target = item
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected member u_1002")
	}
	if target.Status != "banned" {
		t.Fatalf("member status = %q", target.Status)
	}
	if target.MutedUntil == nil || target.MutedUntil.Before(time.Now()) {
		t.Fatalf("muted_until = %v", target.MutedUntil)
	}
}
