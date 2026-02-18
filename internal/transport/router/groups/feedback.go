// internal/transport/router/groups/feedback.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureFeedbackRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	feedbackGroup := r.Group("/feedback")

	// Old: feedbackGroup.Use(middleware.JWTAuthMiddleware(... "admin","user"))
	// => now:
	feedbackGroup.Use(authMiddleware, middleware.RequireRoleMiddleware("admin", "user"))

	// POST /feedback
	feedbackGroup.POST("/", c.FeedbackHandler.CreateFeedback)
}
