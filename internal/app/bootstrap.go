package app

import (
	"context"
	"database/sql"
	"log"
	"time"

	"zcyp-im/internal/auth"
	"zcyp-im/internal/config"
	"zcyp-im/internal/eventbus"
	adminhandler "zcyp-im/internal/handler/admin"
	apihandler "zcyp-im/internal/handler/api"
	imhandler "zcyp-im/internal/handler/im"
	"zcyp-im/internal/model"
	"zcyp-im/internal/repository"
	memoryrepo "zcyp-im/internal/repository/memory"
	mysqlrepo "zcyp-im/internal/repository/mysql"
	"zcyp-im/internal/service"
	mysqlstore "zcyp-im/internal/store/mysql"
	wsgateway "zcyp-im/internal/ws"
)

type Bootstrap struct {
	Config          config.Config
	DB              *sql.DB
	AppService      *service.AppService
	UserService     *service.UserService
	TokenService    *auth.TokenService
	AdminHandler    *adminhandler.AppHandler
	UserHandler     *adminhandler.UserHandler
	APIConversation *apihandler.ConversationHandler
	APIMember       *apihandler.MemberHandler
	APIMessage      *apihandler.MessageHandler
	Connection      *imhandler.ConnectionHandler
	MessageHandler  *imhandler.MessageHandler
	MemberHandler   *imhandler.MemberHandler
	Hub             *wsgateway.Hub
	MessageBus      *eventbus.RedisMessageBus
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
	var messageBroadcaster interface {
		BroadcastMessage(conversationNo string, message model.Message)
	} = hub
	var messageBus *eventbus.RedisMessageBus
	if cfg.Redis.Enabled {
		messageBus = eventbus.NewRedisMessageBus(cfg.Redis)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := messageBus.Ping(ctx)
		cancel()
		if err != nil {
			_ = messageBus.Close()
			if db != nil {
				_ = db.Close()
			}
			return nil, err
		}
		messageBroadcaster = messageBus
		hub.SetPublisher(messageBus)
		log.Printf("bootstrap: message_bus=redis address=%s channel=%s", cfg.Redis.Address, cfg.Redis.Channel)
	} else {
		log.Printf("bootstrap: message_bus=memory")
	}

	return &Bootstrap{
		Config:          cfg,
		DB:              db,
		AppService:      appService,
		UserService:     userService,
		TokenService:    tokenService,
		AdminHandler:    adminhandler.NewAppHandler(appService),
		UserHandler:     adminhandler.NewUserHandler(userService),
		APIConversation: apihandler.NewConversationHandler(imService),
		APIMember:       apihandler.NewMemberHandler(imService),
		APIMessage:      apihandler.NewMessageHandler(imService, messageBroadcaster),
		Connection:      imhandler.NewConnectionHandler(appService, userService, tokenService, hub),
		MessageHandler:  imhandler.NewMessageHandler(imService, messageBroadcaster),
		MemberHandler:   imhandler.NewMemberHandler(imService),
		Hub:             hub,
		MessageBus:      messageBus,
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
