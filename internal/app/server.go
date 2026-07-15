package app

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/auth"
	adminhandler "zcyp-im/internal/handler/admin"
	apihandler "zcyp-im/internal/handler/api"
	imhandler "zcyp-im/internal/handler/im"
	"zcyp-im/internal/response"
	"zcyp-im/internal/service"
)

type Server struct {
	engine    *gin.Engine
	bootstrap *Bootstrap
	host      string
	port      int
}

func NewServer() (*Server, error) {
	bootstrap, err := NewBootstrap()
	if err != nil {
		return nil, err
	}

	engine := gin.Default()
	_ = engine.SetTrustedProxies(nil)
	engine.Use(response.Middleware())
	registerAPIRoutes(
		engine,
		bootstrap.AppService,
		bootstrap.UserService,
		bootstrap.TokenService,
		bootstrap.AdminHandler,
		bootstrap.UserHandler,
		bootstrap.APIConversation,
		bootstrap.APIMember,
		bootstrap.APIMessage,
		bootstrap.Connection,
		bootstrap.MessageHandler,
		bootstrap.MemberHandler,
	)

	return &Server{
		engine:    engine,
		bootstrap: bootstrap,
		host:      bootstrap.Config.HTTP.Host,
		port:      bootstrap.Config.HTTP.Port,
	}, nil
}

func NewWebSocketServer() (*Server, error) {
	bootstrap, err := NewBootstrap()
	if err != nil {
		return nil, err
	}

	engine := gin.Default()
	_ = engine.SetTrustedProxies(nil)
	engine.Use(response.Middleware())
	registerWebSocketRoutes(engine, bootstrap.Connection)

	return &Server{
		engine:    engine,
		bootstrap: bootstrap,
		host:      bootstrap.Config.WebSocket.Host,
		port:      bootstrap.Config.WebSocket.Port,
	}, nil
}

func (s *Server) Run() error {
	defer func() {
		if s.bootstrap.DB != nil {
			_ = s.bootstrap.DB.Close()
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("server: listen addr=%s", addr)
	return s.engine.Run(addr)
}

func registerAPIRoutes(
	engine *gin.Engine,
	appService *service.AppService,
	userService *service.UserService,
	tokenService *auth.TokenService,
	adminHandler *adminhandler.AppHandler,
	userHandler *adminhandler.UserHandler,
	apiConversationHandler *apihandler.ConversationHandler,
	apiMemberHandler *apihandler.MemberHandler,
	apiMessageHandler *apihandler.MessageHandler,
	imHandler *imhandler.ConnectionHandler,
	messageHandler *imhandler.MessageHandler,
	memberHandler *imhandler.MemberHandler,
) {
	engine.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status":  "ok",
			"service": "zcyp-im-api",
		})
	})

	admin := engine.Group("/admin")
	{
		admin.GET("/apps", adminHandler.ListApps)
		admin.POST("/apps", adminHandler.CreateApp)
		admin.GET("/apps/:app_code", adminHandler.GetApp)
		admin.GET("/apps/:app_code/users", userHandler.ListUsers)
		admin.POST("/apps/:app_code/users", userHandler.UpsertUser)
		admin.GET("/apps/:app_code/users/:external_user_id", userHandler.GetUser)
		admin.PATCH("/apps/:app_code/users/:external_user_id/status", userHandler.UpdateUserStatus)
	}

	apiUserHandler := apihandler.NewUserHandler(userService, tokenService)
	api := engine.Group("/api")
	api.Use(apihandler.AppAuthMiddleware(appService))
	{
		api.POST("/users", apiUserHandler.UpsertUser)
		api.GET("/users/:external_user_id", apiUserHandler.GetUser)
		api.PUT("/users/:external_user_id", apiUserHandler.UpdateUser)
		api.POST("/users/:external_user_id/token", apiUserHandler.IssueAccessToken)
		api.POST("/conversations", apiConversationHandler.CreateConversation)
		api.GET("/conversations/:conversation_no/messages", apiConversationHandler.ListMessages)
		api.GET("/conversations/:conversation_no/members", apiConversationHandler.ListMembers)
		api.POST("/conversations/:conversation_no/members", apiMemberHandler.AddMembers)
		api.DELETE("/conversations/:conversation_no/members/:member_user_id", apiMemberHandler.RemoveMember)
		api.POST("/conversations/:conversation_no/all-muted", apiMemberHandler.UpdateConversationAllMuted)
		api.POST("/conversations/:conversation_no/review", apiMemberHandler.UpdateConversationReview)
		api.POST("/conversations/:conversation_no/members/:member_user_id/ban", apiMemberHandler.BanMember)
		api.POST("/conversations/:conversation_no/members/:member_user_id/unban", apiMemberHandler.UnbanMember)
		api.POST("/conversations/:conversation_no/members/:member_user_id/role", apiMemberHandler.UpdateMemberRole)
		api.POST("/conversations/:conversation_no/members/:member_user_id/mic", apiMemberHandler.UpdateMemberMic)
		api.POST("/conversations/:conversation_no/members/:member_user_id/mute", apiMemberHandler.MuteMember)
		api.POST("/conversations/:conversation_no/members/:member_user_id/unmute", apiMemberHandler.UnmuteMember)
		api.POST("/conversations/:conversation_no/messages", apiMessageHandler.SendMessage)
	}

	im := engine.Group("/im")
	{
		im.POST("/auth/token", imHandler.IssueAccessToken)
	}

	imAuthorized := engine.Group("/im")
	imAuthorized.Use(imhandler.AuthMiddleware(tokenService))
	{
		imAuthorized.POST("/conversations", messageHandler.CreateConversation)
		imAuthorized.POST("/conversations/:conversation_no/messages", messageHandler.SendMessage)
		imAuthorized.GET("/conversations/:conversation_no/messages", messageHandler.ListMessages)
		imAuthorized.GET("/conversations/:conversation_no/members", memberHandler.ListMembers)
		imAuthorized.POST("/conversations/:conversation_no/members", memberHandler.AddMembers)
		imAuthorized.DELETE("/conversations/:conversation_no/members/:member_user_id", memberHandler.RemoveMember)
		imAuthorized.POST("/conversations/:conversation_no/join", memberHandler.JoinConversation)
		imAuthorized.POST("/conversations/:conversation_no/leave", memberHandler.LeaveConversation)
		imAuthorized.POST("/conversations/:conversation_no/all-muted", memberHandler.UpdateConversationAllMuted)
		imAuthorized.POST("/conversations/:conversation_no/review", memberHandler.UpdateConversationReview)
		imAuthorized.POST("/conversations/:conversation_no/members/:member_user_id/ban", memberHandler.BanMember)
		imAuthorized.POST("/conversations/:conversation_no/members/:member_user_id/unban", memberHandler.UnbanMember)
		imAuthorized.POST("/conversations/:conversation_no/members/:member_user_id/role", memberHandler.UpdateMemberRole)
		imAuthorized.POST("/conversations/:conversation_no/members/:member_user_id/mic", memberHandler.UpdateMemberMic)
		imAuthorized.POST("/conversations/:conversation_no/members/:member_user_id/mute", memberHandler.MuteMember)
		imAuthorized.POST("/conversations/:conversation_no/members/:member_user_id/unmute", memberHandler.UnmuteMember)
	}
}

func registerWebSocketRoutes(engine *gin.Engine, imHandler *imhandler.ConnectionHandler) {
	engine.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{
			"status":  "ok",
			"service": "zcyp-im-ws",
		})
	})

	engine.GET("/im/connect", imHandler.Connect)
}
