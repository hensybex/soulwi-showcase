// internal/transport/router/groups/chat.go
package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureChatRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	chatGroup := r.Group("/chats")

	// All chat routes require auth
	chatGroup.Use(authMiddleware, middleware.RequireRoleMiddleware("admin", "user"))

	// Support both /chats and /chats/
	chatGroup.GET("", c.ChatHandler.ListChats)
	chatGroup.GET("/", c.ChatHandler.ListChats)

	chatGroup.POST("/", c.ChatHandler.CreateChat)
	chatGroup.GET("/:chat_id", c.ChatHandler.GetChat)
	chatGroup.DELETE("/:chat_id", c.ChatHandler.DeleteChat)
	chatGroup.PATCH("/:chat_id/rename", c.ChatHandler.RenameChat)

	// Message routes
	chatGroup.GET("/:chat_id/messages", c.MessageHandler.ListMessages)
	chatGroup.DELETE("/:chat_id/messages/:message_id", c.MessageHandler.DeleteMessage)

	// SSE
	chatGroup.POST("/:chat_id/sse",
		c.MessageLimitMiddleware.CheckMessageLimit(),
		c.RateLimiterMiddleware.CheckRateLimit(),
		c.SSEHandler.ChatStreamSSE,
	)
	chatGroup.POST("/:chat_id/sse-mock",
		c.MessageLimitMiddleware.CheckMessageLimit(),
		c.RateLimiterMiddleware.CheckRateLimit(),
		c.SSEHandler.MockChatStreamSSE,
	)
}
