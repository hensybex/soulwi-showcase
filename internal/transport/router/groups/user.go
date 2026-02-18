// Файл: internal/transport/router/groups/user.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
	"github.com/hensybex/soulwi_go_back/internal/transport/middleware"
)

func ConfigureUserRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	userGroup := r.Group("/users")

	// Применяем базовую аутентификацию (проверка Firebase токена) ко всей группе.
	userGroup.Use(authMiddleware)

	// --- МАРШРУТЫ, КОТОРЫМ НЕ НУЖНА ПРОВЕРКА РОЛИ В НАШЕЙ БД ---
	// Эти эндпоинты должны быть доступны любому пользователю с валидным Firebase токеном.

	// <<< ИСПРАВЛЕНИЕ: Вызываем настоящий метод RegisterOrUpdateUser >>>
	userGroup.POST("/sync", c.UserHandler.RegisterOrUpdateUser)
	userGroup.POST("/fcm-token", c.UserHandler.UpdateFCMToken)

	// --- МАРШРУТЫ, КОТОРЫМ НУЖЕН СУЩЕСТВУЮЩИЙ ПОЛЬЗОВАТЕЛЬ С РОЛЬЮ ---
	// Этот мидлвар проверяет, что пользователь уже есть в нашей БД и имеет роль.
	authenticated := userGroup.Group("/")
	authenticated.Use(middleware.RequireRoleMiddleware("admin", "user"))
	{
		// <<< ИСПРАВЛЕНИЕ: Вызываем настоящий метод GetCurrentUser >>>
		authenticated.GET("/me", c.UserHandler.GetCurrentUser)

		// <<< ИСПРАВЛЕНИЕ: Для обновления данных профиля также используем RegisterOrUpdateUser >>>
		authenticated.PUT("/me", c.UserHandler.RegisterOrUpdateUser)
		authenticated.DELETE("/me", c.UserHandler.DeleteCurrentUser)
	}
}
