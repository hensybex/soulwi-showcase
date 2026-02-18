// api/internal/transport/router/groups/debug.go
package groups

import (
	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/di"
)

func RegisterDebugRoutes(r *gin.Engine, c *di.Container) {
	g := r.Group("/debug")
	g.POST("/push/by-uid", c.NotifyHandler.SendByUID)
	g.POST("/push/by-token", c.NotifyHandler.SendByToken)
}
