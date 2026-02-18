package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/config"
	"github.com/hensybex/soulwi_go_back/internal/usecase"
)

type CronHandler struct {
	notificationUC usecase.NotificationUsecase
	cfg            *config.Config
}

func NewCronHandler(notificationUC usecase.NotificationUsecase, cfg *config.Config) *CronHandler {
	return &CronHandler{notificationUC: notificationUC, cfg: cfg}
}

func (h *CronHandler) RunDaily(c *gin.Context) {
	if !h.authorize(c) {
		return
	}

	ctx := c.Request.Context()
	current := time.Now().UTC().Hour()

	h.notificationUC.SendDailyNotifications(ctx, 9, current)
	h.notificationUC.SendDailyNotifications(ctx, 21, current)

	c.JSON(http.StatusOK, gin.H{"status": "daily notifications dispatched", "current_utc_hour": current})
}

func (h *CronHandler) RunReengagement(c *gin.Context) {
	if !h.authorize(c) {
		return
	}

	h.notificationUC.SendReengagementNotifications(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"status": "reengagement notifications dispatched"})
}

func (h *CronHandler) authorize(c *gin.Context) bool {
	key := h.cfg.CronAuthKey
	if key == "" {
		return true
	}

	header := c.GetHeader("X-Cron-Key")
	if header == key {
		return true
	}

	if c.Query("key") == key {
		return true
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	return false
}
