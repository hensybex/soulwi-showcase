// File: internal/transport/router/router.go

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"

	// Import the handler package to access the RequestLogger
	"github.com/hensybex/soulwi_go_back/internal/handler"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
	"github.com/hensybex/soulwi_go_back/internal/transport/router/groups"
)

func SetupRouter(container *di.Container) *gin.Engine {
	r := gin.Default()
	r.RedirectTrailingSlash = false

	// Custom logger
	r.Use(handler.RequestLogger())

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(
		container.Config.JWTAccessSecret,
		container.FirebaseAuthClient,
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Grouped routes
	groups.ConfigureAuthRoutes(r, container)
	groups.ConfigureUserRoutes(r, container, authMiddleware)
	groups.ConfigurePromptRoutes(r, container, authMiddleware)
	groups.ConfigureChatRoutes(r, container, authMiddleware)
	groups.ConfigureNoteRoutes(r, container, authMiddleware)
	groups.ConfigureDailyCheckInRoutes(r, container, authMiddleware)
	groups.ConfigureWisePhraseRoutes(r, container, authMiddleware)
	groups.ConfigureTodoRoutes(r, container, authMiddleware)
	groups.ConfigureFeedbackRoutes(r, container, authMiddleware)
	groups.ConfigureAppleRoutes(r, container, authMiddleware)
	groups.ConfigureDashboardRoutes(r, container, authMiddleware)

	if container.Config.EnableDevRoutes {
		groups.RegisterDebugRoutes(r, container)
	}
	groups.RegisterCronRoutes(r, container)

	return r
}
