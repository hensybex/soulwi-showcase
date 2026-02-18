package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
)

func RegisterCronRoutes(r *gin.Engine, c *di.Container) {
	cron := r.Group("/cron")
	cron.POST("/daily", c.CronHandler.RunDaily)
	cron.POST("/reengagement", c.CronHandler.RunReengagement)
}
