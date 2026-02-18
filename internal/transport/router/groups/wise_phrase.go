// internal/transport/router/groups/wise_phrase.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureWisePhraseRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	wpGroup := r.Group("/wise-phrases")
	wpGroup.Use(authMiddleware)

	// Admin-only endpoints
	wpGroup.POST("/generate", middleware.RequireRoleMiddleware("admin"), c.WisePhraseHandler.GenerateBatch)
	wpGroup.GET("/admin", middleware.RequireRoleMiddleware("admin"), c.WisePhraseHandler.ListAdmin)
	wpGroup.DELETE("/:id", middleware.RequireRoleMiddleware("admin"), c.WisePhraseHandler.Delete)

	// Shared admin/user endpoints
	wpGroup.GET("/random", middleware.RequireRoleMiddleware("admin", "user"), c.WisePhraseHandler.GetRandom)
	wpGroup.POST("/:id/toggleLike", middleware.RequireRoleMiddleware("admin", "user"), c.WisePhraseHandler.ToggleLikePhrase)
	wpGroup.POST("/:id/share", middleware.RequireRoleMiddleware("admin", "user"), c.WisePhraseHandler.RecordShare)
	wpGroup.GET("/liked", middleware.RequireRoleMiddleware("admin", "user"), c.WisePhraseHandler.ListLikes)
}
