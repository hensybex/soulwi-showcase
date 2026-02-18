// internal/transport/router/groups/apple.go
package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureAppleRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	appleGroup := r.Group("/apple")

	// Эндпоинт для валидации чеков - требует аутентификации
	appleGroup.POST("/verify-receipt",
		authMiddleware,
		middleware.RequireRoleMiddleware("user", "admin"),
		c.AppleHandler.VerifyReceipt,
	)

	appleGroup.POST("/verify-receipt-test",
		authMiddleware,
		middleware.RequireRoleMiddleware("user", "admin"),
		c.AppleHandler.VerifyReceiptTest,
	)

	// Эндпоинт для server notifications от Apple - без аутентификации
	// (вызывается напрямую серверами Apple)
	appleGroup.POST("/server-notifications", c.AppleHandler.HandleServerNotifications)

	// Дополнительный эндпоинт для проверки статуса подписки
	appleGroup.GET("/subscription-status",
		authMiddleware,
		middleware.RequireRoleMiddleware("user", "admin"),
		c.AppleHandler.GetSubscriptionStatus,
	)
}
