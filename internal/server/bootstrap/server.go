package bootstrap

import (
	"context"
	"log"
	"net/http"

	"github.com/hensybex/soulwi_go_back/internal/config"
	"github.com/hensybex/soulwi_go_back/internal/cron"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/router"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

type Server struct {
	db            *gorm.DB
	cfg           *config.Config
	container     *di.Container
	httpServer    *http.Server
	cronScheduler *cron.CronScheduler
}

func NewServer(db *gorm.DB, cfg *config.Config, container *di.Container) *Server {
	return &Server{
		db:        db,
		cfg:       cfg,
		container: container,
	}
}

func (s *Server) Run() error {
	log.Println("[Server] Run method called.")

	// --- ЗАПУСК CRON ПЛАНИРОВЩИКА ---
	log.Println("[Server] Initializing cron scheduler...")
	s.cronScheduler = cron.NewCronScheduler(
		s.container.SubscriptionUsecase,
		s.container.NotificationUsecase,
	)

	go s.cronScheduler.Start()
	log.Println("[Server] Cron scheduler started in a separate goroutine.")

	// Setup the router
	log.Println("[Server] Setting up router...")
	r := router.SetupRouter(s.container)
	log.Println("[Server] Router setup complete.")

	// Setup CORS middleware
	log.Println("[Server] Setting up CORS...")
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept", "X-Firebase-AppCheck"},
		AllowCredentials: true,
	})
	handler := corsHandler.Handler(r)
	log.Println("[Server] CORS setup complete.")

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    ":" + s.cfg.ApiPort,
		Handler: handler,
	}

	// Start listening
	log.Printf("[Server] Attempting to start HTTP server on port %s...", s.cfg.ApiPort)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("[Server] Failed to start server: %v", err)
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("[Server] Starting graceful shutdown...")

	// Stop cron scheduler
	if s.cronScheduler != nil {
		log.Println("[Server] Stopping cron scheduler...")
		s.cronScheduler.Stop()
	}

	// Shutdown HTTP server
	if s.httpServer != nil {
		log.Println("[Server] Shutting down HTTP server...")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("[Server] Error shutting down HTTP server: %v", err)
			return err
		}
		log.Println("[Server] HTTP server stopped.")
	}

	log.Println("[Server] Graceful shutdown complete.")
	return nil
}
