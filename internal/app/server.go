package app

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"

	"zcyp-im/internal/auth"
	"zcyp-im/internal/config"
	adminhandler "zcyp-im/internal/handler/admin"
	imhandler "zcyp-im/internal/handler/im"
	"zcyp-im/internal/repository"
	memoryrepo "zcyp-im/internal/repository/memory"
	mysqlrepo "zcyp-im/internal/repository/mysql"
	"zcyp-im/internal/service"
	mysqlstore "zcyp-im/internal/store/mysql"
	wsgateway "zcyp-im/internal/ws"
)

type Server struct {
	engine *gin.Engine
	config config.Config
	db     *sql.DB
}

func NewServer() (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	engine := gin.Default()

	stores, db, err := buildRepositories(cfg)
	if err != nil {
		return nil, err
	}

	appService := service.NewAppService(stores.appRepo)
	userService := service.NewUserService(appService, stores.userRepo)
	tokenService := auth.NewTokenService(cfg.JWT)
	imService := service.NewIMService(appService, userService, stores.conversationRepo, stores.memberRepo, stores.messageRepo, cfg.Moderation.BlockedWords)
	hub := wsgateway.NewHub(imService)
	adminHandler := adminhandler.NewAppHandler(appService)
	userHandler := adminhandler.NewUserHandler(userService)
	imHandler := imhandler.NewConnectionHandler(appService, userService, tokenService, hub)
	messageHandler := imhandler.NewMessageHandler(imService, hub)
	memberHandler := imhandler.NewMemberHandler(imService)

	registerRoutes(engine, tokenService, adminHandler, userHandler, imHandler, messageHandler, memberHandler)

	return &Server{
		engine: engine,
		config: cfg,
		db:     db,
	}, nil
}

func (s *Server) Run() error {
	defer func() {
		if s.db != nil {
			_ = s.db.Close()
		}
	}()

	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)
	return s.engine.Run(addr)
}

type repositories struct {
	appRepo          repository.AppRepository
	userRepo         repository.UserRepository
	conversationRepo repository.ConversationRepository
	memberRepo       repository.ConversationMemberRepository
	messageRepo      repository.MessageRepository
}

func buildRepositories(cfg config.Config) (repositories, *sql.DB, error) {
	if !cfg.MySQL.Enabled {
		return repositories{
			appRepo:          memoryrepo.NewAppRepository(),
			userRepo:         memoryrepo.NewUserRepository(),
			conversationRepo: memoryrepo.NewConversationRepository(),
			memberRepo:       memoryrepo.NewConversationMemberRepository(),
			messageRepo:      memoryrepo.NewMessageRepository(),
		}, nil, nil
	}

	db, err := mysqlstore.New(cfg.MySQL)
	if err != nil {
		return repositories{}, nil, err
	}

	return repositories{
		appRepo:          mysqlrepo.NewAppRepository(db),
		userRepo:         mysqlrepo.NewUserRepository(db),
		conversationRepo: mysqlrepo.NewConversationRepository(db),
		memberRepo:       mysqlrepo.NewConversationMemberRepository(db),
		messageRepo:      mysqlrepo.NewMessageRepository(db),
	}, db, nil
}

func registerRoutes(
	engine *gin.Engine,
	tokenService *auth.TokenService,
	adminHandler *adminhandler.AppHandler,
	userHandler *adminhandler.UserHandler,
	imHandler *imhandler.ConnectionHandler,
	messageHandler *imhandler.MessageHandler,
	memberHandler *imhandler.MemberHandler,
) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "zcyp-im",
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

	im := engine.Group("/im")
	{
		im.POST("/auth/token", imHandler.IssueAccessToken)
		im.GET("/connect", imHandler.Connect)
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
