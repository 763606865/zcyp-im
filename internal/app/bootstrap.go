package app

import (
	"database/sql"
	"log"

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

type Bootstrap struct {
	Config         config.Config
	DB             *sql.DB
	TokenService   *auth.TokenService
	AdminHandler   *adminhandler.AppHandler
	UserHandler    *adminhandler.UserHandler
	Connection     *imhandler.ConnectionHandler
	MessageHandler *imhandler.MessageHandler
	MemberHandler  *imhandler.MemberHandler
}

func NewBootstrap() (*Bootstrap, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	if cfg.MySQL.Enabled {
		log.Printf("bootstrap: storage=mysql host=%s port=%d database=%s user=%s", cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database, cfg.MySQL.Username)
	} else {
		log.Printf("bootstrap: storage=memory")
	}

	stores, db, err := buildRepositories(cfg)
	if err != nil {
		return nil, err
	}

	appService := service.NewAppService(stores.appRepo)
	userService := service.NewUserService(appService, stores.userRepo)
	tokenService := auth.NewTokenService(cfg.JWT)
	imService := service.NewIMService(appService, userService, stores.conversationRepo, stores.memberRepo, stores.messageRepo, cfg.Moderation.BlockedWords)
	hub := wsgateway.NewHub(imService)

	return &Bootstrap{
		Config:         cfg,
		DB:             db,
		TokenService:   tokenService,
		AdminHandler:   adminhandler.NewAppHandler(appService),
		UserHandler:    adminhandler.NewUserHandler(userService),
		Connection:     imhandler.NewConnectionHandler(appService, userService, tokenService, hub),
		MessageHandler: imhandler.NewMessageHandler(imService, hub),
		MemberHandler:  imhandler.NewMemberHandler(imService),
	}, nil
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
