// internal/di/di.go
package di

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"github.com/hensybex/soulwi_go_back/internal/config"
	"github.com/hensybex/soulwi_go_back/internal/handler"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/service"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type Container struct {
	Config                     *config.Config
	FirebaseAuthClient         *firebaseAuth.Client
	UserRepository             repository.UserRepository
	SubscriptionRepository     repository.SubscriptionRepository
	AppleService               service.AppleService
	NotificationService        service.NotificationService
	ChatUsecase                usecase.ChatUsecase
	MessageUsecase             usecase.MessageUsecase
	PromptUsecase              usecase.PromptUsecase
	BasePromptUsecase          usecase.BasePromptUsecase
	NoteUsecase                usecase.NoteUsecase
	DailyCheckInUsecase        usecase.DailyCheckInUsecase
	WisePhraseUsecase          usecase.WisePhraseUsecase
	TodoUsecase                usecase.TodoUsecase
	FeedbackUsecase            usecase.FeedbackUsecase
	PromptVersionUsecase       usecase.PromptVersionUsecase
	SubgroupsAndPromptsUsecase usecase.SubgroupsAndPromptsUsecase
	UserUsecase                usecase.UserUsecase
	SubscriptionUsecase        usecase.SubscriptionUsecase
	NotificationUsecase        usecase.NotificationUsecase
	MessageLimitMiddleware     *middleware.MessageLimitMiddleware
	PromptHandler              *handler.PromptHandler
	ChatHandler                *handler.ChatHandler
	MessageHandler             *handler.MessageHandler
	SSEHandler                 *handler.SSEHandler
	AuthHandler                *handler.AuthHandler
	BasePromptHandler          *handler.BasePromptHandler
	NoteHandler                *handler.NoteHandler
	DailyCheckInHandler        *handler.DailyCheckInHandler
	WisePhraseHandler          *handler.WisePhraseHandler
	TodoHandler                *handler.TodoHandler
	FeedbackHandler            *handler.FeedbackHandler
	PromptVersionHandler       *handler.PromptVersionHandler
	SubgroupsAndPromptsHandler *handler.SubgroupsAndPromptsHandler
	UserHandler                *handler.UserHandler
	AppleHandler               *handler.AppleHandler
	NotifyHandler              *handler.NotifyHandler
	CronHandler                *handler.CronHandler
	RateLimiterMiddleware      *middleware.RateLimiterMiddleware
	DashboardHandler           *handler.DashboardHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	// Firebase init prefers inline JSON (Cloud Run secrets) and falls back to file when present.
	var firebaseOpts []option.ClientOption
	var fbAuthClient *firebaseAuth.Client
	var messagingClient *messaging.Client

	switch {
	case cfg.FirebaseCredsJSON != "":
		firebaseOpts = append(firebaseOpts, option.WithCredentialsJSON([]byte(cfg.FirebaseCredsJSON)))
	case cfg.FirebaseCredsFile != "":
		if _, err := os.Stat(cfg.FirebaseCredsFile); err != nil {
			log.Printf("[DI] FIREBASE_CREDS_FILE is set but not readable: %v", err)
		} else {
			firebaseOpts = append(firebaseOpts, option.WithCredentialsFile(cfg.FirebaseCredsFile))
		}
	default:
		log.Println("[DI] Firebase credentials are not configured. Firebase-dependent endpoints will return 503.")
	}

	if len(firebaseOpts) > 0 {
		firebaseApp, err := firebase.NewApp(context.Background(), nil, firebaseOpts...)
		if err != nil {
			log.Printf("[DI] Failed to initialize Firebase app: %v", err)
		} else {
			authClient, err := firebaseApp.Auth(context.Background())
			if err != nil {
				log.Printf("[DI] Failed to initialize Firebase Auth: %v", err)
			} else {
				fbAuthClient = authClient
			}

			msgClient, err := firebaseApp.Messaging(context.Background())
			if err != nil {
				log.Printf("[DI] Failed to initialize Firebase Messaging: %v", err)
			} else {
				messagingClient = msgClient
			}
		}
	}

	// Repositories
	promptRepo := repository.NewPromptRepo(db)
	chatRepo := repository.NewChatRepo(db)
	messageRepo := repository.NewMessageRepo(db)
	basePromptRepo := repository.NewBasePromptRepo(db)
	noteRepo := repository.NewNoteRepo(db)
	dciRepo := repository.NewDailyCheckInRepo(db)
	wpRepo := repository.NewWisePhraseRepo(db)
	wpLikeRepo := repository.NewWisePhraseLikeRepo(db)
	wpShareRepo := repository.NewWisePhraseShareRepo(db)
	todoRepo := repository.NewTodoRepo(db)
	feedbackRepo := repository.NewFeedbackRepo(db)
	promptVersionRepo := repository.NewPromptVersionRepo(db)
	userRepo := repository.NewUserRepo(db)
	subscriptionRepo := repository.NewSubscriptionRepo(db)
	dashboardRepo := repository.NewDashboardRepository(db)

	// Services
	appleService := service.NewAppleService(
		cfg.AppleAppSharedSecret,
		cfg.AppleBundleID,
	)
	notificationService := service.NewNotificationService(messagingClient)

	// Usecases
	chatUC := usecase.NewChatUsecase(chatRepo, promptRepo)
	messageUC := usecase.NewMessageUsecase(messageRepo, chatRepo)
	aiUsecase := usecase.NewAIUsecase(chatRepo, messageRepo, promptRepo, basePromptRepo, cfg.OpenAIKey)
	basePromptUC := usecase.NewBasePromptUsecase(basePromptRepo)
	noteUC := usecase.NewNoteUsecase(noteRepo)
	promptVersionUC := usecase.NewPromptVersionUsecase(promptVersionRepo)
	promptUC := usecase.NewPromptUsecase(promptRepo, basePromptRepo, promptVersionUC)
	subgroupsAndPromptsUC := usecase.NewSubgroupsAndPromptsUsecase(promptRepo)
	todoUC := usecase.NewTodoUsecase(todoRepo)
	userUC := usecase.NewUserUsecase(
		userRepo,
		noteUC,
		todoUC,
		noteRepo,
		todoRepo,
		dciRepo,
		chatRepo,
		messageRepo,
		wpLikeRepo,
		wpShareRepo,
		subscriptionRepo,
	)
	dciUC := usecase.NewDailyCheckInUsecase(dciRepo)
	wpUC := usecase.NewWisePhraseUsecase(wpRepo, wpLikeRepo, wpShareRepo, promptRepo, aiUsecase)
	feedbackUC := usecase.NewFeedbackUsecase(feedbackRepo)
	subscriptionUC := usecase.NewSubscriptionUsecase(subscriptionRepo, userRepo, appleService)
	notificationUC := usecase.NewNotificationUsecase(userRepo, notificationService)
	dashboardUC := usecase.NewDashboardUsecase(dashboardRepo)

	// Middleware
	messageLimitMiddleware := middleware.NewMessageLimitMiddleware(userRepo)
	rateLimit := rate.Limit(30.0 / 60.0)
	burst := 15
	rateLimiter := middleware.NewRateLimiterMiddleware(rateLimit, burst)

	// Handlers
	promptHandler := handler.NewPromptHandler(promptUC)
	chatHandler := handler.NewChatHandler(chatUC)
	messageHandler := handler.NewMessageHandler(messageUC, chatUC)
	sseHandler := handler.NewSSEHandler(aiUsecase, messageUC, messageRepo, chatUC, messageLimitMiddleware)
	authHandler := handler.NewAuthHandler(cfg.JWTAccessSecret, fbAuthClient, userRepo)
	basePromptHandler := handler.NewBasePromptHandler(basePromptUC)
	noteHandler := handler.NewNoteHandler(noteUC)
	dciHandler := handler.NewDailyCheckInHandler(dciUC, userUC)
	wpHandler := handler.NewWisePhraseHandler(wpUC)
	todoHandler := handler.NewTodoHandler(todoUC)
	feedbackHandler := handler.NewFeedbackHandler(feedbackUC)
	promptVersionHandler := handler.NewPromptVersionHandler(promptVersionUC)
	subgroupsAndPromptsHandler := handler.NewSubgroupsAndPromptsHandler(subgroupsAndPromptsUC)
	userHandler := handler.NewUserHandler(userUC, fbAuthClient)
	appleHandler := handler.NewAppleHandler(subscriptionUC, fbAuthClient)
	// NOTE: pass the concrete NotificationService, not the usecase
	notifyHandler := handler.NewNotifyHandler(userRepo, notificationService)
	cronHandler := handler.NewCronHandler(notificationUC, cfg)
	dashboardHandler := handler.NewDashboardHandler(dashboardUC)

	return &Container{
		Config:                     cfg,
		FirebaseAuthClient:         fbAuthClient,
		UserRepository:             userRepo,
		SubscriptionRepository:     subscriptionRepo,
		AppleService:               appleService,
		NotificationService:        notificationService,
		ChatUsecase:                chatUC,
		MessageUsecase:             messageUC,
		PromptUsecase:              promptUC,
		BasePromptUsecase:          basePromptUC,
		NoteUsecase:                noteUC,
		DailyCheckInUsecase:        dciUC,
		WisePhraseUsecase:          wpUC,
		TodoUsecase:                todoUC,
		FeedbackUsecase:            feedbackUC,
		PromptVersionUsecase:       promptVersionUC,
		SubgroupsAndPromptsUsecase: subgroupsAndPromptsUC,
		UserUsecase:                userUC,
		SubscriptionUsecase:        subscriptionUC,
		NotificationUsecase:        notificationUC,
		MessageLimitMiddleware:     messageLimitMiddleware,
		PromptHandler:              promptHandler,
		ChatHandler:                chatHandler,
		MessageHandler:             messageHandler,
		SSEHandler:                 sseHandler,
		AuthHandler:                authHandler,
		BasePromptHandler:          basePromptHandler,
		NoteHandler:                noteHandler,
		DailyCheckInHandler:        dciHandler,
		WisePhraseHandler:          wpHandler,
		TodoHandler:                todoHandler,
		FeedbackHandler:            feedbackHandler,
		PromptVersionHandler:       promptVersionHandler,
		SubgroupsAndPromptsHandler: subgroupsAndPromptsHandler,
		UserHandler:                userHandler,
		AppleHandler:               appleHandler,
		NotifyHandler:              notifyHandler,
		CronHandler:                cronHandler,
		RateLimiterMiddleware:      rateLimiter,
		DashboardHandler:           dashboardHandler,
	}
}
