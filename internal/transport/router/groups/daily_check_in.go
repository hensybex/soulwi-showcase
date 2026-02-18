// internal/transport/router/groups/daily_check_in.go

package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
)

func ConfigureDailyCheckInRoutes(r *gin.Engine, c *di.Container, authMiddleware gin.HandlerFunc) {
	dciGroup := r.Group("/checkins")

	dciGroup.Use(authMiddleware)

	// Все маршруты ниже теперь защищены только проверкой Firebase токена.
	dciGroup.GET("", c.DailyCheckInHandler.ListCheckIns)

	// <<< FIX IS HERE >>>
	// Change the path from "/" to "" to register POST /checkins instead of POST /checkins/
	dciGroup.POST("", c.DailyCheckInHandler.CreateCheckIn)

	dciGroup.GET("/:id", c.DailyCheckInHandler.GetCheckIn)
	dciGroup.PUT("/:id", c.DailyCheckInHandler.UpdateCheckIn)
	dciGroup.DELETE("/:id", c.DailyCheckInHandler.DeleteCheckIn)
	dciGroup.GET("/status", c.DailyCheckInHandler.CheckInStatus)
}
