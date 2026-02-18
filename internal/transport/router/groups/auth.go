// internal/transport/router/groups/auth.go (обновленная версия)
package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
)

func ConfigureAuthRoutes(r *gin.Engine, c *di.Container) {
	authGroup := r.Group("/auth")

	// Public authentication endpoints.
	authGroup.POST("/apple-signin", c.AuthHandler.AppleSignIn)

	// Development-only endpoints are opt-in.
	if c.Config.EnableDevRoutes {
		authGroup.GET("/test-token", c.AuthHandler.GetTestToken)
		authGroup.POST("/register-test-user", c.AuthHandler.RegisterTestUser)
		authGroup.POST("/get-fb-token", c.AuthHandler.GetFirebaseToken)
		authGroup.POST("/test-full-login", c.AuthHandler.TestFullLogin)
	}
}
