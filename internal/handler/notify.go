// api/internal/handler/notify.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hensybex/soulwi_go_back/internal/repository"
	"github.com/hensybex/soulwi_go_back/internal/service"
)

type NotifyHandler struct {
	users repository.UserRepository
	send  service.NotificationService
}

func NewNotifyHandler(ur repository.UserRepository, ns service.NotificationService) *NotifyHandler {
	return &NotifyHandler{users: ur, send: ns}
}

type byUID struct {
	FirebaseUID string `json:"firebase_uid" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Body        string `json:"body" binding:"required"`
	Route       string `json:"route"`
}

type byToken struct {
	Token string `json:"token" binding:"required"`
	Title string `json:"title" binding:"required"`
	Body  string `json:"body" binding:"required"`
	Route string `json:"route"`
}

func (h *NotifyHandler) SendByUID(c *gin.Context) {
	var req byUID
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, err := h.users.GetByFirebaseUID(c.Request.Context(), req.FirebaseUID)
	if err != nil || u == nil || u.FCMToken == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found or no token"})
		return
	}
	if err := h.send.SendRawToken(c.Request.Context(), u.FCMToken, req.Title, req.Body, map[string]string{"route": req.Route}); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *NotifyHandler) SendByToken(c *gin.Context) {
	var req byToken
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.send.SendRawToken(c.Request.Context(), req.Token, req.Title, req.Body, map[string]string{"route": req.Route}); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
