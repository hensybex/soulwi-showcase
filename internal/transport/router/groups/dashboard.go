// internal/transport/router/groups/dashboard.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
)

func ConfigureDashboardRoutes(r *gin.Engine, container *di.Container, authMiddleware gin.HandlerFunc) {
	api := r.Group("/api")
	api.Use(authMiddleware)
	{
		api.GET("/dashboard", container.DashboardHandler.GetDashboardData)
		api.GET("/dashboard/infographic", container.DashboardHandler.GetInfographicData)
	}
}
